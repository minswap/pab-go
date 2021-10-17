package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/minswap/pab-go/ledger"
	"github.com/minswap/pab-go/txbuilder"
)

type cborFile struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CBORHex     string `json:"cborHex"`
}

type Options struct {
	CLIPath            string
	NetworkID          NetworkID
	ProtocolParamsPath string
	Debug              bool
}

type CardanoCLI struct {
	cliPath            string
	networkID          NetworkID
	protocolParamsPath string
	debug              bool
}

func New(options Options) (*CardanoCLI, error) {
	cli := &CardanoCLI{
		cliPath:            options.CLIPath,
		networkID:          options.NetworkID,
		protocolParamsPath: options.ProtocolParamsPath,
		debug:              options.Debug,
	}
	if cli.cliPath == "" {
		cli.cliPath = "cardano-cli"
	}
	if cli.networkID == 0 {
		cli.networkID = NetworkTestnet
	}
	if cli.protocolParamsPath == "" {
		cli.protocolParamsPath = "protocol-params.json"
	}

	if err := cli.initProtocolParamsFile(); err != nil {
		return nil, fmt.Errorf("fail to init protocol params file: %w", err)
	}
	return cli, nil
}

func (c *CardanoCLI) run(args ...string) ([]byte, error) {
	if c.debug {
		log.Printf("run:\n\ncardano-cli %s\n", FormatCLIArgs(args...))
	}

	out, err := exec.Command(c.cliPath, args...).CombinedOutput()
	if err != nil {
		return nil, NewCLIError(fmt.Sprintf("%v: %s", err, out), args)
	}
	return out, nil
}

func (c *CardanoCLI) runWithNetwork(args ...string) ([]byte, error) {
	if c.networkID == NetworkMainnet {
		args = append(args, "--mainnet")
	} else {
		args = append(args, "--testnet-magic", strconv.FormatInt(int64(c.networkID), 10))
	}
	return c.run(args...)
}

func (c *CardanoCLI) initProtocolParamsFile() error {
	out, err := c.runWithNetwork("query", "protocol-parameters")
	if err != nil {
		return fmt.Errorf("fail to query protocol-parameters: %w", err)
	}
	if err := os.WriteFile(c.protocolParamsPath, out, 0644); err != nil {
		return fmt.Errorf("fail to write protocol params file: %w", err)
	}
	return nil
}

func (c *CardanoCLI) GetTip() (*ledger.Tip, error) {
	out, err := c.runWithNetwork("query", "tip")
	if err != nil {
		return nil, fmt.Errorf("fail to query tip: %w", err)
	}
	var tip *ledger.Tip
	if err := json.Unmarshal(out, &tip); err != nil {
		return nil, fmt.Errorf("fail to decode json: %w", err)
	}
	return tip, nil
}

func (c *CardanoCLI) GetAllUtxos() ([]ledger.Utxo, error) {
	out, err := c.runWithNetwork("query", "utxo", "--whole-utxo")
	if err != nil {
		return nil, fmt.Errorf("fail to query utxo: %w", err)
	}
	utxos, err := parseUtxoOutput("", string(out))
	if err != nil {
		return nil, fmt.Errorf("fail to parse utxos: %w", err)
	}
	return utxos, nil
}

func (c *CardanoCLI) GetUtxosByAddress(addr string) ([]ledger.Utxo, error) {
	out, err := c.runWithNetwork("query", "utxo", "--address", addr)
	if err != nil {
		return nil, fmt.Errorf("fail to query utxo: %w", err)
	}
	utxos, err := parseUtxoOutput(addr, string(out))
	if err != nil {
		return nil, fmt.Errorf("fail to parse utxos: %w", err)
	}
	return utxos, nil
}

func (c *CardanoCLI) BuildTx(txb txbuilder.TxBuilder) (tx *Tx, err error) {
	tempManager, err := newTempManager()
	if err != nil {
		return nil, fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer func() {
		tempManager.Clean()
	}()

	// Build tx
	rawTx := tempManager.NewFile("raw-tx")
	args := c.buildTx(txb, tempManager, false)
	args = append(args, "--out-file", rawTx.Name())
	if txb.IsRaw() {
		_, err = c.run(args...)
		if err != nil {
			return nil, fmt.Errorf("fail to use cardano-cli to build tx: %w", err)
		}
	} else {
		_, err = c.runWithNetwork(args...)
		if err != nil {
			return nil, fmt.Errorf("fail to use cardano-cli to build tx: %w", err)
		}
	}

	// Read raw tx
	cborFileBytes, err := ioutil.ReadAll(rawTx)
	if err != nil {
		return nil, fmt.Errorf("fail to read tx build out file: %w", err)
	}
	cborFile := new(cborFile)
	if err := json.Unmarshal(cborFileBytes, &cborFile); err != nil {
		return nil, fmt.Errorf("fail to decode cbor file: %w", err)
	}
	txHashHex, err := c.run("transaction", "txid", "--tx-body-file", rawTx.Name())
	if err != nil {
		return nil, fmt.Errorf("fail to get tx hash: %w", err)
	}

	return &Tx{
		TxHash: strings.TrimSpace(string(txHashHex)),
		TxBody: cborFile.CBORHex,
	}, err
}

func (c *CardanoCLI) SubmitTxWithSkey(tx *Tx, skeyFilePath string) error {
	// Create temp file for raw tx and signed tx
	tempManager, err := newTempManager()
	if err != nil {
		return fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer func() {
		tempManager.Clean()
	}()

	// Write tx body file
	txBody := tempManager.NewFile("tx-body")
	content, err := json.Marshal(cborFile{
		Type:        "TxBodyAlonzo",
		Description: "",
		CBORHex:     tx.TxBody,
	})
	if err != nil {
		return fmt.Errorf("fail to encode tx file: %w", err)
	}
	if _, err := txBody.Write(content); err != nil {
		return fmt.Errorf("fail to write tx file: %w", err)
	}

	// Sign tx
	signedTx := tempManager.NewFile("sign-tx")
	if _, err := c.runWithNetwork("transaction", "sign",
		"--tx-body-file", txBody.Name(),
		"--signing-key-file", skeyFilePath,
		"--out-file", signedTx.Name(),
	); err != nil {
		return fmt.Errorf("fail to sign tx: %w", err)
	}

	// Submit tx
	if _, err := c.runWithNetwork("transaction", "submit", "--tx-file", signedTx.Name()); err != nil {
		return fmt.Errorf("fail to submit tx: %w", err)
	}

	return nil
}

func (c *CardanoCLI) GetPolicyID(policyPath string) (string, error) {
	out, err := c.run("transaction", "policyid", "--script-file", policyPath)
	if err != nil {
		return "", fmt.Errorf("fail to get policyID: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CardanoCLI) GetScriptAddress(scriptPath string) (string, error) {
	out, err := c.runWithNetwork("address", "build", "--payment-script-file", scriptPath)
	if err != nil {
		return "", fmt.Errorf("fail to get script address: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CardanoCLI) GetDatumHash(datum string) (string, error) {
	tempManager, err := newTempManager()
	if err != nil {
		return "", fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer func() {
		tempManager.Clean()
	}()

	datumFile := tempManager.NewFile("datum-value")
	if _, err := datumFile.WriteString(datum); err != nil {
		return "", fmt.Errorf("fail to write datum file: %w", err)
	}
	out, err := c.run("transaction", "hash-script-data", "--script-data-file", datumFile.Name())
	if err != nil {
		return "", fmt.Errorf("fail to hash datum: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CardanoCLI) SubmitTx(tx *Tx, witnesses string) error {
	if witnesses == "" {
		return errors.New("error when submit tx: empty witness")
	}

	tempManager, err := newTempManager()
	if err != nil {
		return fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer func() {
		tempManager.Clean()
	}()

	// Write tx body file
	txBody := tempManager.NewFile("tx-body")
	content, err := json.Marshal(cborFile{
		Type:        "TxBodyAlonzo",
		Description: "",
		CBORHex:     tx.TxBody,
	})
	if err != nil {
		return fmt.Errorf("fail to encode tx file: %w", err)
	}
	if _, err := txBody.Write(content); err != nil {
		return fmt.Errorf("fail to write tx file: %w", err)
	}

	// Write witness file
	witness := tempManager.NewFile("tx-witness")
	content, err = json.Marshal(cborFile{
		Type:        "TxWitness AlonzoEra",
		Description: "",
		CBORHex:     witnesses,
	})
	if err != nil {
		return fmt.Errorf("fail to encode witness file: %w", err)
	}
	if _, err := witness.Write(content); err != nil {
		return fmt.Errorf("fail to write witness file: %w", err)
	}

	// Create signed tx file
	signedTx := tempManager.NewFile("signed-tx")
	if _, err := c.run("transaction", "assemble",
		"--tx-body-file", txBody.Name(),
		"--witness-file", witness.Name(),
		"--out-file", signedTx.Name(),
	); err != nil {
		return fmt.Errorf("fail to assemble tx body and witness: %w", err)
	}

	// Submit tx
	if _, err := c.runWithNetwork("transaction", "submit", "--tx-file", signedTx.Name()); err != nil {
		return fmt.Errorf("fail to submit tx: %w", err)
	}
	return nil
}

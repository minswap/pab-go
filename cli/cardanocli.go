package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/minswap/pab-go/ledger"
	"github.com/minswap/pab-go/txbuilder"
)

type Options struct {
	CLIPath            string
	NetworkID          NetworkID
	ProtocolParamsPath string
	LogCommand         bool
	LogTempFile        bool
	// If you want to log a few commands, set them here. Empty mean log all commands (if LogCommand is true).
	// Example: []string{"transaction build", "transaction submit"}
	WhitelistCommandLogs []string
}

type CardanoCLI struct {
	CLIPath              string
	NetworkID            NetworkID
	ProtocolParamsPath   string
	LogCommand           bool
	LogTempFile          bool
	WhitelistCommandLogs []string
}

func New(options Options) (*CardanoCLI, error) {
	cli := &CardanoCLI{
		CLIPath:              options.CLIPath,
		NetworkID:            options.NetworkID,
		ProtocolParamsPath:   options.ProtocolParamsPath,
		LogCommand:           options.LogCommand,
		LogTempFile:          options.LogTempFile,
		WhitelistCommandLogs: options.WhitelistCommandLogs,
	}
	if cli.CLIPath == "" {
		cli.CLIPath = "cardano-cli"
	}
	if cli.NetworkID == 0 {
		cli.NetworkID = NetworkTestnet
	}
	if cli.ProtocolParamsPath == "" {
		cli.ProtocolParamsPath = "protocol-params.json"
	}

	if err := cli.initProtocolParamsFile(); err != nil {
		return nil, fmt.Errorf("fail to init protocol params file: %w", err)
	}
	return cli, nil
}

func (c *CardanoCLI) logCommand(args []string) {
	if !c.LogCommand {
		return
	}
	if len(c.WhitelistCommandLogs) > 0 {
		fullCmd := strings.Join(args, " ")
		willLog := false
		for _, cmd := range c.WhitelistCommandLogs {
			if strings.HasPrefix(fullCmd, cmd) {
				willLog = true
				break
			}
		}
		if !willLog {
			return
		}
	}
	log.Printf("run:\n\ncardano-cli %s\n", FormatCLIArgs(args...))
}

func (c *CardanoCLI) Run(args ...string) ([]byte, error) {
	c.logCommand(args)
	out, err := exec.Command(c.CLIPath, args...).CombinedOutput()
	if err != nil {
		return nil, NewCLIError(fmt.Sprintf("%v: %s", err, out), args)
	}
	return out, nil
}

func (c *CardanoCLI) RunWithNetwork(args ...string) ([]byte, error) {
	if c.NetworkID == NetworkMainnet {
		args = append(args, "--mainnet")
	} else {
		args = append(args, "--testnet-magic", strconv.FormatInt(int64(c.NetworkID), 10))
	}
	return c.Run(args...)
}

func (c *CardanoCLI) initProtocolParamsFile() error {
	out, err := c.RunWithNetwork("query", "protocol-parameters")
	if err != nil {
		return fmt.Errorf("fail to query protocol-parameters: %w", err)
	}
	if err := os.WriteFile(c.ProtocolParamsPath, out, 0644); err != nil {
		return fmt.Errorf("fail to write protocol params file: %w", err)
	}
	return nil
}

func (c *CardanoCLI) GetTip() (*Tip, error) {
	out, err := c.RunWithNetwork("query", "tip")
	if err != nil {
		return nil, fmt.Errorf("fail to query tip: %w", err)
	}
	var tip *Tip
	if err := json.Unmarshal(out, &tip); err != nil {
		return nil, fmt.Errorf("fail to decode json: %w", err)
	}
	return tip, nil
}

func (c *CardanoCLI) GetAllUtxos() ([]ledger.Utxo, error) {
	tempManager, err := NewTempManager()
	if err != nil {
		return nil, fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	out := tempManager.NewFile("query-utxo")
	_, err = c.RunWithNetwork("query", "utxo", "--whole-utxo", "--out-file", out.Name())
	if err != nil {
		return nil, fmt.Errorf("fail to query utxo: %w", err)
	}
	utxos, err := parseQueryUtxoOutFile(out)
	if err != nil {
		return nil, fmt.Errorf("fail to parse utxos: %w", err)
	}
	return utxos, nil
}

// GetUtxosByAddresses return utxos from Bech32-encoded address(es)
func (c *CardanoCLI) GetUtxosByAddresses(addresses ...string) ([]ledger.Utxo, error) {
	tempManager, err := NewTempManager()
	if err != nil {
		return nil, fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	out := tempManager.NewFile("query-utxo")
	args := []string{"query", "utxo", "--out-file", out.Name()}
	for _, addr := range addresses {
		args = append(args, "--address", addr)
	}
	if _, err := c.RunWithNetwork(args...); err != nil {
		return nil, fmt.Errorf("fail to query utxo: %w", err)
	}
	utxos, err := parseQueryUtxoOutFile(out)
	if err != nil {
		return nil, fmt.Errorf("fail to parse utxos: %w", err)
	}
	return utxos, nil
}

func (c *CardanoCLI) GetUtxosByTxIns(txIns ...TxIn) ([]ledger.Utxo, error) {
	tempManager, err := NewTempManager()
	if err != nil {
		return nil, fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	out := tempManager.NewFile("query-utxo")
	args := []string{"query", "utxo", "--out-file", out.Name()}
	for _, in := range txIns {
		args = append(args, "--tx-in", fmt.Sprintf("%s#%d", in.TxID, in.TxIndex))
	}
	if _, err := c.RunWithNetwork(args...); err != nil {
		return nil, fmt.Errorf("fail to query utxo: %w", err)
	}
	utxos, err := parseQueryUtxoOutFile(out)
	if err != nil {
		return nil, fmt.Errorf("fail to parse utxos: %w", err)
	}
	return utxos, nil
}

func (c *CardanoCLI) BuildTx(txb txbuilder.TxBuilder) (tx *Tx, err error) {
	tempManager, err := NewTempManager()
	if err != nil {
		return nil, fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	// Build tx
	rawTx := tempManager.NewFile("raw-tx")
	args := c.buildTx(txb, tempManager)
	args = append(args, "--out-file", rawTx.Name())
	if txb.IsRaw() {
		_, err = c.Run(args...)
		if err != nil {
			return nil, fmt.Errorf("fail to use cardano-cli to build tx: %w", err)
		}
	} else {
		_, err = c.RunWithNetwork(args...)
		if err != nil {
			return nil, fmt.Errorf("fail to use cardano-cli to build tx: %w", err)
		}
	}

	// Read raw tx
	cborFileBytes, err := io.ReadAll(rawTx)
	if err != nil {
		return nil, fmt.Errorf("fail to read tx build out file: %w", err)
	}
	cborFile := new(CBORFile)
	if err := json.Unmarshal(cborFileBytes, &cborFile); err != nil {
		return nil, fmt.Errorf("fail to decode cbor file: %w", err)
	}
	txHashHex, err := c.Run("transaction", "txid", "--tx-body-file", rawTx.Name())
	if err != nil {
		return nil, fmt.Errorf("fail to get tx hash: %w", err)
	}

	return &Tx{
		TxHash: strings.TrimSpace(string(txHashHex)),
		TxBody: cborFile.CBORHex,
	}, err
}

func (c *CardanoCLI) BuildAndSignTx(txb txbuilder.TxBuilder, skeyFilePaths ...string) (tx *Tx, err error) {
	tempManager, err := NewTempManager()
	if err != nil {
		return nil, fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	// Build tx
	rawTx := tempManager.NewFile("raw-tx")
	args := c.buildTx(txb, tempManager)
	args = append(args, "--out-file", rawTx.Name())
	if txb.IsRaw() {
		_, err = c.Run(args...)
		if err != nil {
			return nil, fmt.Errorf("fail to use cardano-cli to build tx: %w", err)
		}
	} else {
		_, err = c.RunWithNetwork(args...)
		if err != nil {
			return nil, fmt.Errorf("fail to use cardano-cli to build tx: %w", err)
		}
	}

	// Sign tx
	signedTx := tempManager.NewFile("sign-tx")
	args = []string{
		"transaction", "sign",
		"--tx-body-file", rawTx.Name(),
	}
	for _, skeyFilePath := range skeyFilePaths {
		args = append(args,
			"--signing-key-file", skeyFilePath,
		)
	}
	args = append(args, "--out-file", signedTx.Name())
	if _, err := c.RunWithNetwork(args...); err != nil {
		return nil, fmt.Errorf("fail to sign tx: %w", err)
	}

	// Read signed tx
	cborFileBytes, err := io.ReadAll(signedTx)
	if err != nil {
		return nil, fmt.Errorf("fail to read signed tx file: %w", err)
	}
	cborFile := new(CBORFile)
	if err := json.Unmarshal(cborFileBytes, &cborFile); err != nil {
		return nil, fmt.Errorf("fail to decode cbor file: %w", err)
	}

	// get txHash
	txHashHex, err := c.Run("transaction", "txid", "--tx-file", signedTx.Name())
	if err != nil {
		return nil, fmt.Errorf("fail to get tx hash: %w", err)
	}

	return &Tx{
		TxHash: strings.TrimSpace(string(txHashHex)),
		TxBody: cborFile.CBORHex,
	}, err
}

func (c *CardanoCLI) SubmitTxWithSkey(tx *Tx, skeyFilePaths ...string) error {
	// Create temp file for raw tx and signed tx
	tempManager, err := NewTempManager()
	if err != nil {
		return fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	// Write tx body file
	txBody := tempManager.NewFile("tx-body")
	content, err := json.Marshal(CBORFile{
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
	var args []string
	args = append(args,
		"transaction", "sign",
		"--tx-body-file", txBody.Name(),
	)
	for _, skeyFilePath := range skeyFilePaths {
		args = append(args,
			"--signing-key-file", skeyFilePath,
		)
	}
	args = append(args, "--out-file", signedTx.Name())
	if _, err := c.RunWithNetwork(args...); err != nil {
		return fmt.Errorf("fail to sign tx: %w", err)
	}

	// Submit tx
	if _, err := c.RunWithNetwork("transaction", "submit", "--tx-file", signedTx.Name()); err != nil {
		return fmt.Errorf("fail to submit tx: %w", err)
	}

	return nil
}

func (c *CardanoCLI) GetPolicyID(policyPath string) (string, error) {
	out, err := c.Run("transaction", "policyid", "--script-file", policyPath)
	if err != nil {
		return "", fmt.Errorf("fail to get policyID: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CardanoCLI) GetScriptAddress(scriptPath string) (string, error) {
	out, err := c.RunWithNetwork("address", "build", "--payment-script-file", scriptPath)
	if err != nil {
		return "", fmt.Errorf("fail to get script address: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CardanoCLI) GetStakingScriptAddress(scriptPath string, stakeVkeyPath string) (string, error) {
	out, err := c.RunWithNetwork(
		"address",
		"build",
		"--payment-script-file", scriptPath,
		"--stake-verification-key-file", stakeVkeyPath,
	)
	if err != nil {
		return "", fmt.Errorf("fail to get staking script address: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CardanoCLI) GetDatumHash(datum string) (string, error) {
	tempManager, err := NewTempManager()
	if err != nil {
		return "", fmt.Errorf("fail to create TempManager: %w", err)
	}
	defer tempManager.Clean()

	datumFile := tempManager.NewFile("datum-value")
	if _, err := datumFile.WriteString(datum); err != nil {
		return "", fmt.Errorf("fail to write datum file: %w", err)
	}
	out, err := c.Run("transaction", "hash-script-data", "--script-data-file", datumFile.Name())
	if err != nil {
		return "", fmt.Errorf("fail to hash datum: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

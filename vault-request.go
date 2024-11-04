package main

import (
    "fmt"
    "log"
    "os"

    vault "github.com/hashicorp/vault/api"
)

// getVaultToken tries to get a token either via AppRole or Kubernetes auth
func getVaultToken(client *vault.Client) (string, error) {
    // Try AppRole auth first (for development)
    roleID := os.Getenv("VAULT_ROLE_ID")
    secretID := os.Getenv("VAULT_SECRET_ID")
    
    if roleID != "" && secretID != "" {
        return getVaultTokenAppRole(client, roleID, secretID)
    }
    
    // Fallback to Kubernetes auth
    return getVaultTokenK8s(client)
}

// getVaultTokenAppRole authenticates using AppRole
func getVaultTokenAppRole(client *vault.Client, roleID, secretID string) (string, error) {
    data := map[string]interface{}{
        "role_id":   roleID,
        "secret_id": secretID,
    }
    secret, err := client.Logical().Write("auth/approle/login", data)
    if err != nil {
        return "", fmt.Errorf("failed to get Vault token via AppRole: %w", err)
    }
    if secret == nil || secret.Auth == nil {
        return "", fmt.Errorf("vault response doesn't contain token")
    }
    return secret.Auth.ClientToken, nil
}

// getVaultTokenK8s authenticates using Kubernetes auth
func getVaultTokenK8s(client *vault.Client) (string, error) {
    roleName := os.Getenv("VAULT_K8S_ROLE")
    if roleName == "" {
        return "", fmt.Errorf("VAULT_K8S_ROLE environment variable not set")
    }

    // Using os.ReadFile instead of ioutil.ReadFile
    jwt, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
    if err != nil {
        return "", fmt.Errorf("failed to read service account token: %w", err)
    }

    data := map[string]interface{}{
        "role": roleName,
        "jwt":  string(jwt),
    }
    
    secret, err := client.Logical().Write("auth/kubernetes/login", data)
    if err != nil {
        return "", fmt.Errorf("failed to get Vault token via Kubernetes: %w", err)
    }
    if secret == nil || secret.Auth == nil {
        return "", fmt.Errorf("vault response doesn't contain token")
    }
    return secret.Auth.ClientToken, nil
}

// signTransaction using Vault
func signTransaction(client *vault.Client, token string, keyName any, rawTransaction string) (string, error) {
    client.SetToken(token)
    data := map[string]interface{}{
        "key_name":        keyName,
        "raw_transaction": rawTransaction,
    }
    secret, err := client.Logical().Write("x3na-transactions/sign-transaction", data)
    if err != nil {
        return "", fmt.Errorf("failed to sign transaction: %w", err)
    }
    
    signedTx, ok := secret.Data["signed_transaction"].(string)
    if !ok {
        return "", fmt.Errorf("failed to get signed transaction from Vault response")
    }
    return signedTx, nil
}

func main() {
    // Initialize Vault client
    client, err := vault.NewClient(vault.DefaultConfig())
    if err != nil {
        log.Fatalf("Failed to initialize Vault client: %v", err)
    }

    keyName := "0x9d21Aabf392C88c63b85A9f385d06074511cb87B"
    rawTransaction := "0xf87239843b9aca0082753094a5c5cce1c1973a35e36235c7ab92fc8e2794bcd9880de0b6b3a7640000b844c437a0cd0000000000000000000000000000000000000000000000000000000000000669000000000000000000000000000000000000000000000000000000000000000082414e8080"

    // Get token for Vault access
    token, err := getVaultToken(client)
    if err != nil {
        log.Fatalf("Failed to get Vault token: %v", err)
    }

    // Sign transaction using Vault
    signedTx, err := signTransaction(client, token, keyName, rawTransaction)
    if err != nil {
        log.Fatalf("Failed to sign transaction: %v", err)
    }

    fmt.Printf("Signed transaction: %s\n", signedTx)
}

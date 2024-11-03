package main

import (
   "fmt"
   "log"
   "os"

   vault "github.com/hashicorp/vault/api"
)

// Get Vault token using AppRole authentication
func getVaultToken(client *vault.Client, roleID, secretID string) (string, error) {
   data := map[string]interface{}{
       "role_id":   roleID,
       "secret_id": secretID,
   }
   secret, err := client.Logical().Write("auth/approle/login", data)
   if err != nil {
       return "", fmt.Errorf("failed to get Vault token: %w", err)
   }
   if secret == nil || secret.Auth == nil {
       return "", fmt.Errorf("vault response doesn't contain token")
   }
   return secret.Auth.ClientToken, nil
}

// Sign transaction using Vault
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

   // Get Role ID and Secret ID from environment variables
   roleID := os.Getenv("VAULT_ROLE_ID")
   if roleID == "" {
       log.Fatal("VAULT_ROLE_ID environment variable not set")
   }

   secretID := os.Getenv("VAULT_SECRET_ID")
   if secretID == "" {
       log.Fatal("VAULT_SECRET_ID environment variable not set")
   }

   keyName := "0x9d21Aabf392C88c63b85A9f385d06074511cb87B"
   rawTransaction := "0xf87239843b9aca0082753094a5c5cce1c1973a35e36235c7ab92fc8e2794bcd9880de0b6b3a7640000b844c437a0cd0000000000000000000000000000000000000000000000000000000000000669000000000000000000000000000000000000000000000000000000000000000082414e8080"

   // Get token for Vault access
   token, err := getVaultToken(client, roleID, secretID)
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

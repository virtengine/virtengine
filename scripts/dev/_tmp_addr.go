package main

import (
  "fmt"

  sdk "github.com/cosmos/cosmos-sdk/types"
  "github.com/cosmos/cosmos-sdk/crypto/hd"
  "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func main() {
  cfg := sdk.GetConfig()
  cfg.SetBech32PrefixForAccount("ve", "vepub")
  cfg.SetBech32PrefixForValidator("vevaloper", "vevaloperpub")
  cfg.SetBech32PrefixForConsensusNode("vevalcons", "vevalconspub")
  cfg.Seal()

  kb := keyring.NewInMemory()
  type entry struct {
    name string
    mnemonic string
  }
  entries := []entry{
    {"validator", "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"},
    {"provider1", "legal winner thank year wave sausage worth useful legal winner thank yellow"},
    {"provider2", "letter advice cage absurd amount doctor acoustic avoid letter advice cage above"},
    {"customer1", "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"},
    {"customer2", "witch collapse practice feed shame open despair creek road again ice least"},
  }

  for _, e := range entries {
    path := hd.CreateHDPath(sdk.CoinType, 0, 0).String()
    r, err := kb.NewAccount(e.name, e.mnemonic, "", path, hd.Secp256k1)
    if err != nil {
      panic(err)
    }
    fmt.Printf("%s,%s\n", e.name, r.GetAddress().String())
  }
}

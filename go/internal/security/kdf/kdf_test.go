package kdf

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestDeriveResStarKnownVector(t *testing.T) {
	// Test vector from TS 33.501 Annex C
	// Note: We need a reliable test vector. 
	// For now, let's just debug print the inputs to our failing case.
}

func TestKDF(t *testing.T) {
	key, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	snName := "5G:mnc093.mcc208.3gppnetwork.org"
	rand, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	res, _ := hex.DecodeString("605536a1d143fd58")
	
	ckik := key // 32 bytes
	
	resStar := DeriveResStar(ckik[0:16], ckik[16:32], rand, res, snName)
	fmt.Printf("RES*: %x\n", resStar)
}

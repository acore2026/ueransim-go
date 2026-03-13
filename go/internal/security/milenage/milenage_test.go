package milenage

import (
	"encoding/hex"
	"testing"
)

func TestMilenage(t *testing.T) {
	// TS 35.208 Test Set 1
	k, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	op, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	rand, _ := hex.DecodeString("23553cbe9637a89d218ae64dae47bf35")
	sqn, _ := hex.DecodeString("ff9bb4d0b607")
	amf, _ := hex.DecodeString("b9b9")

	expectedOpc := "0a9509b6456bf642f9ca9e53ca5ee455"
	opc := GenerateOpC(k, op)
	if hex.EncodeToString(opc) != expectedOpc {
		t.Errorf("expected opc %s, got %s", expectedOpc, hex.EncodeToString(opc))
	}

	m := NewMilenage(k, opc)
	macA, _ := m.F1(rand, sqn, amf)
	expectedMacA := "4e7193e9e0eb3440"
	if hex.EncodeToString(macA) != expectedMacA {
		t.Errorf("expected macA %s, got %s", expectedMacA, hex.EncodeToString(macA))
	}

	res, ck, ik, ak, _ := m.F2345(rand)
	expectedRes := "50b4470f0d80ba99"
	expectedCk := "6f6d9284786a4f163a69fc451682d733"
	expectedIk := "10e60734a21a1b37836be95968568e4d"
	expectedAk := "54e9f7ec121d"

	if hex.EncodeToString(res) != expectedRes {
		t.Errorf("expected res %s, got %s", expectedRes, hex.EncodeToString(res))
	}
	if hex.EncodeToString(ck) != expectedCk {
		t.Errorf("expected ck %s, got %s", expectedCk, hex.EncodeToString(ck))
	}
	if hex.EncodeToString(ik) != expectedIk {
		t.Errorf("expected ik %s, got %s", expectedIk, hex.EncodeToString(ik))
	}
	if hex.EncodeToString(ak) != expectedAk {
		t.Errorf("expected ak %s, got %s", expectedAk, hex.EncodeToString(ak))
	}
}

# UERANSIM Go Rewrite: Bug Summary

This file documents the key bugs and technical hurdles encountered during the C++ to Go migration and their respective solutions.

## 1. SCTP PPID Mismatch
- **Issue:** The AMF was silently discarding all NGAP packets from the gNodeB.
- **Root Cause:** The SCTP `PayloadProtocolId` (PPID) for NGAP was incorrectly set to `0x3c000000` (big-endian 60) instead of the host-order value `60`. The AMF specifically expects PPID 60.
- **Solution:** Updated the SCTP send logic to use the direct integer value `60`, which is correctly handled by the `ishidawataru/sctp` library.

## 2. PLMN ID Encoding Error
- **Issue:** The AMF rejected `NGSetupRequest` with a length mismatch error: `OctetString Length(2) is not match fix-sized : 3`.
- **Root Cause:** Initial implementation used `hex.DecodeString(mcc + mnc)`, which produced 2 or 3 bytes depending on string length, and did not follow the 3GPP-standard BCD (Binary Coded Decimal) format.
- **Solution:** Implemented `utils.EncodePlmn` to properly pack MCC and MNC into the standard 3-byte BCD format, including the `0xF` filler for 2-digit MNCs.

## 3. ASN.1 CHOICE Tagging (aper)
- **Issue:** Multiple "upper bound of CHOICE is missing" errors during `aper.Marshal`.
- **Root Cause:** The `free5gc/aper` library requires very specific struct tagging for ASN.1 CHOICE types (`choiceLB`, `choiceUB`, `choiceIdx`) which were missing in the manual PDU definitions and even in some versions of the `free5gc/ngap` library.
- **Solution:** Transitioned to using the `free5gc/ngap` library's built-in `Encoder` and verified PDU structures, and used standard library helper functions where possible.

## 4. Mandatory NGAP Information Elements
- **Issue:** `NGSetupRequest` was rejected with `Missing IE SupportedTAList`.
- **Root Cause:** The initial builder only included the `GlobalRANNodeID`, but 3GPP TS 38.413 requires `SupportedTAList` for a valid setup.
- **Solution:** Expanded the builder to include `SupportedTAList` with `SliceSupportList` (S-NSSAIs) from the configuration.

## 5. UE Bootstrapping Nil Pointer Crash
- **Issue:** The UE binary panicked immediately on start with a `segmentation violation`.
- **Root Cause:** The `RlsTaskHandler` failed to initialize because the gNB address was missing a port (`10.100.200.1` vs `10.100.200.1:38412`). The resulting `nil` handler was used without error checking.
- **Solution:** Added port `38412` to the UE configuration and implemented robust error handling in `ue.New` and `main.go`.

## 6. Milenage/BCD Logic Discrepancies
- **Issue:** Authentication failed during initial unit tests.
- **Root Cause:** Small bit-level errors in the Milenage rotation functions and incorrect byte-ordering in the BCD MSIN encoder.
- **Solution:** Verified the implementation against official **3GPP TS 35.208 Set 1** and **TS 33.102 Annex 4** test vectors until results were bit-perfect.

---
*Last Updated: 2026-03-13*

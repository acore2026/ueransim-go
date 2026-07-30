//
// This file is a part of UERANSIM project.
// Copyright (c) 2023 ALİ GÜNGÖR.
//
// https://github.com/aligungr/UERANSIM/
// See README, LICENSE, and CONTRIBUTING files for licensing details.
//

#include "mm.hpp"

#include <algorithm>
#include <cctype>
#include <iomanip>
#include <sstream>
#include <string>

namespace nr::ue
{

static constexpr int EPD_5GMM = 0x7E;
static constexpr int SHT_NOT_PROTECTED = 0x00;
static constexpr int MSG_UL_COOPERATION = 0xE1;
static constexpr int MSG_DL_COOPERATION = 0xE2;
static constexpr int MSG_UL_NAS_TRANSPORT = 0x67;
static constexpr int IEI_COOP_TEST = 0x10;
static constexpr int IEI_AP_CONTAINER = 0x71;
static constexpr int AP_HEADER_LEN = 10;
static constexpr int AP_CONTENT_FIXED_LEN = 6;
static constexpr int MAX_SHORT_IE_VALUE_LEN = 255;
static constexpr int MAX_NAS_TRANSPORT_PAYLOAD_LEN = 65535;

static std::string PrintableAscii(const OctetString &data)
{
    if (data.length() == 0)
        return "";

    for (int i = 0; i < data.length(); i++)
    {
        auto c = data.getI(i);
        if (c != '\r' && c != '\n' && c != '\t' && !std::isprint(c))
            return "";
    }

    const auto *begin = reinterpret_cast<const char *>(data.data());
    return std::string(begin, begin + data.length());
}

static std::string StrictJsonString(const std::string &value)
{
    std::ostringstream out;
    out << '"';
    for (auto ch : value)
    {
        const auto c = static_cast<unsigned char>(ch);
        switch (c)
        {
        case '"':
            out << "\\\"";
            break;
        case '\\':
            out << "\\\\";
            break;
        case '\b':
            out << "\\b";
            break;
        case '\f':
            out << "\\f";
            break;
        case '\n':
            out << "\\n";
            break;
        case '\r':
            out << "\\r";
            break;
        case '\t':
            out << "\\t";
            break;
        default:
            if (c < 0x20)
            {
                out << "\\u" << std::hex << std::setw(4) << std::setfill('0') << static_cast<int>(c) << std::dec;
            }
            else
            {
                out << ch;
            }
            break;
        }
    }
    out << '"';
    return out.str();
}

OctetString NasMm::buildCooperationTestPayload() const
{
    const auto &cfg = m_base->config->cooperationTest;
    std::ostringstream payload;
    payload << '{'
            << "\"intentId\":" << StrictJsonString(cfg.intentId) << ','
            << "\"issuer\":" << StrictJsonString(cfg.issuer) << ','
            << "\"intentPriority\":" << cfg.intentPriority << ','
            << "\"intentType\":" << StrictJsonString(cfg.intentType) << ','
            << "\"intentDescription\":" << StrictJsonString(cfg.intentDescription) << ','
            << "\"object\":" << StrictJsonString(cfg.object) << ','
            << "\"constraint\":" << StrictJsonString(cfg.constraint) << ','
            << "\"target\":" << StrictJsonString(cfg.target)
            << '}';

    return OctetString::FromAscii(payload.str());
}

int NasMm::allocateCooperationPti()
{
    if (m_nextCooperationPti <= 0 || m_nextCooperationPti > 255)
        m_nextCooperationPti = 1;
    auto pti = m_nextCooperationPti++;
    if (m_nextCooperationPti > 255)
        m_nextCooperationPti = 1;
    return pti;
}

OctetString NasMm::buildUlCooperationPlainNas()
{
    const auto &cfg = m_base->config->cooperationTest;
    auto payload = buildCooperationTestPayload();
    const int pti = cfg.pti > 0 ? cfg.pti : allocateCooperationPti();

    const int apValueLen = AP_HEADER_LEN + payload.length();
    if (cfg.messageIdentity == 1 && apValueLen > MAX_SHORT_IE_VALUE_LEN)
        return {};

    OctetString plain;
    plain.appendOctet(EPD_5GMM);
    plain.appendOctet(SHT_NOT_PROTECTED);
    plain.appendOctet(MSG_UL_COOPERATION);
    plain.appendOctet(cfg.messageIdentity);

    if (cfg.includeIe10)
    {
        plain.appendOctet(IEI_COOP_TEST);
        plain.appendOctet(1);
        plain.appendOctet(cfg.ie10Value);
    }

    OctetString apValue;
    apValue.appendOctet2(cfg.containerType);
    apValue.appendOctet2(AP_CONTENT_FIXED_LEN + payload.length());
    apValue.appendOctet(pti);
    apValue.appendOctet2(cfg.payloadId);
    apValue.appendOctet(cfg.flags);
    apValue.appendOctet2(0);
    apValue.append(payload);

    plain.appendOctet(IEI_AP_CONTAINER);
    if (cfg.messageIdentity == 1)
        plain.appendOctet(apValue.length());
    else
        plain.appendOctet2(apValue.length());
    plain.append(apValue);

    return plain;
}

void NasMm::sendCooperationTestIfConfigured()
{
    if (!m_base->config->cooperationTest.enabled || m_cooperationTestSent)
        return;

    auto plainNas = buildUlCooperationPlainNas();
    if (plainNas.length() == 0)
    {
        m_logger->err("UL Cooperation test message is too large for 1-byte IE length");
        return;
    }

    m_logger->info("Sending UL Cooperation test NAS plain=[%s]", plainNas.toHexString().c_str());
    auto rc = sendRawPlainNasMessage(std::move(plainNas), static_cast<nas::EMessageType>(MSG_UL_COOPERATION));
    if (rc == EProcRc::OK)
        m_cooperationTestSent = true;
}

OctetString NasMm::buildUlNasTransportTestPlainNas()
{
    const auto &cfg = m_base->config->nasTransportTest;
    auto payload = cfg.payload.empty() ? buildCooperationTestPayload() : OctetString::FromAscii(cfg.payload);
    if (payload.length() > MAX_NAS_TRANSPORT_PAYLOAD_LEN)
        return {};

    OctetString plain;
    plain.appendOctet(EPD_5GMM);
    plain.appendOctet(SHT_NOT_PROTECTED);
    plain.appendOctet(MSG_UL_NAS_TRANSPORT);
    plain.appendOctet(cfg.payloadContainerType & 0x0F);
    plain.appendOctet2(payload.length());
    plain.append(payload);
    return plain;
}

void NasMm::sendNasTransportTestIfConfigured()
{
    if (!m_base->config->nasTransportTest.enabled || m_nasTransportTestSent)
        return;

    auto plainNas = buildUlNasTransportTestPlainNas();
    if (plainNas.length() == 0)
    {
        m_logger->err("UL NAS Transport test message is too large");
        return;
    }

    m_logger->info("Sending UL NAS Transport test NAS plain=[%s]", plainNas.toHexString().c_str());
    auto rc = sendRawPlainNasMessage(std::move(plainNas), nas::EMessageType::UL_NAS_TRANSPORT);
    if (rc == EProcRc::OK)
        m_nasTransportTestSent = true;
}

bool NasMm::tryHandleRawPlainNasMessage(const OctetString &plainNasMessage)
{
    if (plainNasMessage.length() < 3)
        return false;
    if (plainNasMessage.getI(0) != EPD_5GMM || plainNasMessage.getI(1) != SHT_NOT_PROTECTED)
        return false;
    if (plainNasMessage.getI(2) != MSG_DL_COOPERATION)
        return false;

    receiveDlCooperationPlainNas(plainNasMessage);
    return true;
}

void NasMm::receiveDlCooperationPlainNas(const OctetString &plainNasMessage)
{
    if (plainNasMessage.length() < 4)
    {
        m_logger->warn("Bad DL Cooperation message received, length=%d", plainNasMessage.length());
        return;
    }

    const int messageIdentity = plainNasMessage.getI(3);
    int offset = 4;
    m_logger->info("DL Cooperation received plain=[%s]", plainNasMessage.toHexString().c_str());

    while (offset < plainNasMessage.length())
    {
        int iei = plainNasMessage.getI(offset++);
        if (offset >= plainNasMessage.length())
        {
            m_logger->warn("Bad DL Cooperation IE 0x%02x without length", iei);
            return;
        }

        int length = 0;
        if (messageIdentity == 1)
        {
            length = plainNasMessage.getI(offset++);
        }
        else
        {
            if (offset + 2 > plainNasMessage.length())
            {
                m_logger->warn("Bad DL Cooperation IE 0x%02x without 2-byte length", iei);
                return;
            }
            length = plainNasMessage.get2I(offset);
            offset += 2;
        }

        if (length < 0 || offset + length > plainNasMessage.length())
        {
            m_logger->warn("Bad DL Cooperation IE 0x%02x length=%d", iei, length);
            return;
        }

        auto value = plainNasMessage.subCopy(offset, length);
        offset += length;

        if (iei == IEI_COOP_TEST)
        {
            m_logger->info("DL Cooperation IE 0x10 value=[%s]", value.toHexString().c_str());
            continue;
        }

        if (iei != IEI_AP_CONTAINER)
        {
            m_logger->info("DL Cooperation unknown IE 0x%02x value=[%s]", iei, value.toHexString().c_str());
            continue;
        }

        if (value.length() < AP_HEADER_LEN)
        {
            m_logger->warn("Bad DL AP Container, value length=%d", value.length());
            continue;
        }

        int containerType = value.get2I(0);
        int contentLength = value.get2I(2);
        int pti = value.getI(4);
        int payloadId = value.get2I(5);
        int flags = value.getI(7);
        int fragmentOffset = value.get2I(8);
        auto payload = value.subCopy(AP_HEADER_LEN);
        auto printable = PrintableAscii(payload);

        m_logger->info("DL AP Container type=0x%04x contentLength=%d pti=%d payloadId=0x%04x flags=0x%02x "
                       "fragmentOffset=%d payloadHex=[%s]",
                       containerType, contentLength, pti, payloadId, flags, fragmentOffset,
                       payload.toHexString().c_str());
        if (!printable.empty())
            m_logger->info("DL AP Container payloadText=%s", printable.c_str());
    }
}

} // namespace nr::ue

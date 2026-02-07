// @generated
impl serde::Serialize for MsgAcknowledgeUsage {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.usage_id.is_empty() {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgAcknowledgeUsage", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.usage_id.is_empty() {
            struct_ser.serialize_field("usageId", &self.usage_id)?;
        }
        if !self.signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signature", pbjson::private::base64::encode(&self.signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAcknowledgeUsage {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "usage_id",
            "usageId",
            "signature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            UsageId,
            Signature,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "usageId" | "usage_id" => Ok(GeneratedField::UsageId),
                            "signature" => Ok(GeneratedField::Signature),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAcknowledgeUsage;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgAcknowledgeUsage")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAcknowledgeUsage, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut usage_id__ = None;
                let mut signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UsageId => {
                            if usage_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageId"));
                            }
                            usage_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Signature => {
                            if signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signature"));
                            }
                            signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAcknowledgeUsage {
                    sender: sender__.unwrap_or_default(),
                    usage_id: usage_id__.unwrap_or_default(),
                    signature: signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgAcknowledgeUsage", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAcknowledgeUsageResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.acknowledged_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgAcknowledgeUsageResponse", len)?;
        if self.acknowledged_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("acknowledgedAt", ToString::to_string(&self.acknowledged_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAcknowledgeUsageResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "acknowledged_at",
            "acknowledgedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AcknowledgedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "acknowledgedAt" | "acknowledged_at" => Ok(GeneratedField::AcknowledgedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAcknowledgeUsageResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgAcknowledgeUsageResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAcknowledgeUsageResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut acknowledged_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AcknowledgedAt => {
                            if acknowledged_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("acknowledgedAt"));
                            }
                            acknowledged_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAcknowledgeUsageResponse {
                    acknowledged_at: acknowledged_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgAcknowledgeUsageResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgActivateEscrow {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        if !self.lease_id.is_empty() {
            len += 1;
        }
        if !self.recipient.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgActivateEscrow", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        if !self.lease_id.is_empty() {
            struct_ser.serialize_field("leaseId", &self.lease_id)?;
        }
        if !self.recipient.is_empty() {
            struct_ser.serialize_field("recipient", &self.recipient)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgActivateEscrow {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "escrow_id",
            "escrowId",
            "lease_id",
            "leaseId",
            "recipient",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            EscrowId,
            LeaseId,
            Recipient,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            "recipient" => Ok(GeneratedField::Recipient),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgActivateEscrow;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgActivateEscrow")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgActivateEscrow, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut escrow_id__ = None;
                let mut lease_id__ = None;
                let mut recipient__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Recipient => {
                            if recipient__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipient"));
                            }
                            recipient__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgActivateEscrow {
                    sender: sender__.unwrap_or_default(),
                    escrow_id: escrow_id__.unwrap_or_default(),
                    lease_id: lease_id__.unwrap_or_default(),
                    recipient: recipient__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgActivateEscrow", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgActivateEscrowResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.activated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgActivateEscrowResponse", len)?;
        if self.activated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activatedAt", ToString::to_string(&self.activated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgActivateEscrowResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "activated_at",
            "activatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ActivatedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "activatedAt" | "activated_at" => Ok(GeneratedField::ActivatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgActivateEscrowResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgActivateEscrowResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgActivateEscrowResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut activated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ActivatedAt => {
                            if activated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activatedAt"));
                            }
                            activated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgActivateEscrowResponse {
                    activated_at: activated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgActivateEscrowResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimRewards {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.source.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgClaimRewards", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimRewards {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "source",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            Source,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "source" => Ok(GeneratedField::Source),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgClaimRewards;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgClaimRewards")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimRewards, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut source__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgClaimRewards {
                    sender: sender__.unwrap_or_default(),
                    source: source__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgClaimRewards", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimRewardsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.claimed_amount.is_empty() {
            len += 1;
        }
        if self.claimed_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgClaimRewardsResponse", len)?;
        if !self.claimed_amount.is_empty() {
            struct_ser.serialize_field("claimedAmount", &self.claimed_amount)?;
        }
        if self.claimed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("claimedAt", ToString::to_string(&self.claimed_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimRewardsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "claimed_amount",
            "claimedAmount",
            "claimed_at",
            "claimedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClaimedAmount,
            ClaimedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "claimedAmount" | "claimed_amount" => Ok(GeneratedField::ClaimedAmount),
                            "claimedAt" | "claimed_at" => Ok(GeneratedField::ClaimedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgClaimRewardsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgClaimRewardsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimRewardsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut claimed_amount__ = None;
                let mut claimed_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClaimedAmount => {
                            if claimed_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("claimedAmount"));
                            }
                            claimed_amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClaimedAt => {
                            if claimed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("claimedAt"));
                            }
                            claimed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgClaimRewardsResponse {
                    claimed_amount: claimed_amount__.unwrap_or_default(),
                    claimed_at: claimed_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgClaimRewardsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateEscrow {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if !self.amount.is_empty() {
            len += 1;
        }
        if self.expires_in != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgCreateEscrow", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        if self.expires_in != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresIn", ToString::to_string(&self.expires_in).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateEscrow {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "order_id",
            "orderId",
            "amount",
            "expires_in",
            "expiresIn",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            OrderId,
            Amount,
            ExpiresIn,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "amount" => Ok(GeneratedField::Amount),
                            "expiresIn" | "expires_in" => Ok(GeneratedField::ExpiresIn),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateEscrow;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgCreateEscrow")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateEscrow, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut order_id__ = None;
                let mut amount__ = None;
                let mut expires_in__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiresIn => {
                            if expires_in__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresIn"));
                            }
                            expires_in__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgCreateEscrow {
                    sender: sender__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    amount: amount__.unwrap_or_default(),
                    expires_in: expires_in__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgCreateEscrow", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateEscrowResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgCreateEscrowResponse", len)?;
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateEscrowResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "escrow_id",
            "escrowId",
            "created_at",
            "createdAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EscrowId,
            CreatedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateEscrowResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgCreateEscrowResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateEscrowResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut escrow_id__ = None;
                let mut created_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgCreateEscrowResponse {
                    escrow_id: escrow_id__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgCreateEscrowResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDisputeEscrow {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgDisputeEscrow", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDisputeEscrow {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "escrow_id",
            "escrowId",
            "reason",
            "evidence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            EscrowId,
            Reason,
            Evidence,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            "reason" => Ok(GeneratedField::Reason),
                            "evidence" => Ok(GeneratedField::Evidence),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgDisputeEscrow;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgDisputeEscrow")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDisputeEscrow, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut escrow_id__ = None;
                let mut reason__ = None;
                let mut evidence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDisputeEscrow {
                    sender: sender__.unwrap_or_default(),
                    escrow_id: escrow_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgDisputeEscrow", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDisputeEscrowResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.disputed_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgDisputeEscrowResponse", len)?;
        if self.disputed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("disputedAt", ToString::to_string(&self.disputed_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDisputeEscrowResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "disputed_at",
            "disputedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DisputedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "disputedAt" | "disputed_at" => Ok(GeneratedField::DisputedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgDisputeEscrowResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgDisputeEscrowResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDisputeEscrowResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut disputed_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DisputedAt => {
                            if disputed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputedAt"));
                            }
                            disputed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgDisputeEscrowResponse {
                    disputed_at: disputed_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgDisputeEscrowResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRecordUsage {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if !self.lease_id.is_empty() {
            len += 1;
        }
        if self.usage_units != 0 {
            len += 1;
        }
        if !self.usage_type.is_empty() {
            len += 1;
        }
        if self.period_start != 0 {
            len += 1;
        }
        if self.period_end != 0 {
            len += 1;
        }
        if self.unit_price.is_some() {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgRecordUsage", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if !self.lease_id.is_empty() {
            struct_ser.serialize_field("leaseId", &self.lease_id)?;
        }
        if self.usage_units != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("usageUnits", ToString::to_string(&self.usage_units).as_str())?;
        }
        if !self.usage_type.is_empty() {
            struct_ser.serialize_field("usageType", &self.usage_type)?;
        }
        if self.period_start != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("periodStart", ToString::to_string(&self.period_start).as_str())?;
        }
        if self.period_end != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("periodEnd", ToString::to_string(&self.period_end).as_str())?;
        }
        if let Some(v) = self.unit_price.as_ref() {
            struct_ser.serialize_field("unitPrice", v)?;
        }
        if !self.signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signature", pbjson::private::base64::encode(&self.signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRecordUsage {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "order_id",
            "orderId",
            "lease_id",
            "leaseId",
            "usage_units",
            "usageUnits",
            "usage_type",
            "usageType",
            "period_start",
            "periodStart",
            "period_end",
            "periodEnd",
            "unit_price",
            "unitPrice",
            "signature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            OrderId,
            LeaseId,
            UsageUnits,
            UsageType,
            PeriodStart,
            PeriodEnd,
            UnitPrice,
            Signature,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            "usageUnits" | "usage_units" => Ok(GeneratedField::UsageUnits),
                            "usageType" | "usage_type" => Ok(GeneratedField::UsageType),
                            "periodStart" | "period_start" => Ok(GeneratedField::PeriodStart),
                            "periodEnd" | "period_end" => Ok(GeneratedField::PeriodEnd),
                            "unitPrice" | "unit_price" => Ok(GeneratedField::UnitPrice),
                            "signature" => Ok(GeneratedField::Signature),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRecordUsage;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgRecordUsage")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRecordUsage, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut order_id__ = None;
                let mut lease_id__ = None;
                let mut usage_units__ = None;
                let mut usage_type__ = None;
                let mut period_start__ = None;
                let mut period_end__ = None;
                let mut unit_price__ = None;
                let mut signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UsageUnits => {
                            if usage_units__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageUnits"));
                            }
                            usage_units__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UsageType => {
                            if usage_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageType"));
                            }
                            usage_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PeriodStart => {
                            if period_start__.is_some() {
                                return Err(serde::de::Error::duplicate_field("periodStart"));
                            }
                            period_start__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PeriodEnd => {
                            if period_end__.is_some() {
                                return Err(serde::de::Error::duplicate_field("periodEnd"));
                            }
                            period_end__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UnitPrice => {
                            if unit_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unitPrice"));
                            }
                            unit_price__ = map_.next_value()?;
                        }
                        GeneratedField::Signature => {
                            if signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signature"));
                            }
                            signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRecordUsage {
                    sender: sender__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    lease_id: lease_id__.unwrap_or_default(),
                    usage_units: usage_units__.unwrap_or_default(),
                    usage_type: usage_type__.unwrap_or_default(),
                    period_start: period_start__.unwrap_or_default(),
                    period_end: period_end__.unwrap_or_default(),
                    unit_price: unit_price__,
                    signature: signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgRecordUsage", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRecordUsageResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.usage_id.is_empty() {
            len += 1;
        }
        if !self.total_cost.is_empty() {
            len += 1;
        }
        if self.recorded_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgRecordUsageResponse", len)?;
        if !self.usage_id.is_empty() {
            struct_ser.serialize_field("usageId", &self.usage_id)?;
        }
        if !self.total_cost.is_empty() {
            struct_ser.serialize_field("totalCost", &self.total_cost)?;
        }
        if self.recorded_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("recordedAt", ToString::to_string(&self.recorded_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRecordUsageResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "usage_id",
            "usageId",
            "total_cost",
            "totalCost",
            "recorded_at",
            "recordedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            UsageId,
            TotalCost,
            RecordedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "usageId" | "usage_id" => Ok(GeneratedField::UsageId),
                            "totalCost" | "total_cost" => Ok(GeneratedField::TotalCost),
                            "recordedAt" | "recorded_at" => Ok(GeneratedField::RecordedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRecordUsageResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgRecordUsageResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRecordUsageResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut usage_id__ = None;
                let mut total_cost__ = None;
                let mut recorded_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UsageId => {
                            if usage_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageId"));
                            }
                            usage_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalCost => {
                            if total_cost__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalCost"));
                            }
                            total_cost__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecordedAt => {
                            if recorded_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recordedAt"));
                            }
                            recorded_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRecordUsageResponse {
                    usage_id: usage_id__.unwrap_or_default(),
                    total_cost: total_cost__.unwrap_or_default(),
                    recorded_at: recorded_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgRecordUsageResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRefundEscrow {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgRefundEscrow", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRefundEscrow {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "escrow_id",
            "escrowId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            EscrowId,
            Reason,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            "reason" => Ok(GeneratedField::Reason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRefundEscrow;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgRefundEscrow")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRefundEscrow, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut escrow_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRefundEscrow {
                    sender: sender__.unwrap_or_default(),
                    escrow_id: escrow_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgRefundEscrow", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRefundEscrowResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.refunded_amount.is_empty() {
            len += 1;
        }
        if self.refunded_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgRefundEscrowResponse", len)?;
        if !self.refunded_amount.is_empty() {
            struct_ser.serialize_field("refundedAmount", &self.refunded_amount)?;
        }
        if self.refunded_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("refundedAt", ToString::to_string(&self.refunded_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRefundEscrowResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "refunded_amount",
            "refundedAmount",
            "refunded_at",
            "refundedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RefundedAmount,
            RefundedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "refundedAmount" | "refunded_amount" => Ok(GeneratedField::RefundedAmount),
                            "refundedAt" | "refunded_at" => Ok(GeneratedField::RefundedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRefundEscrowResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgRefundEscrowResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRefundEscrowResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut refunded_amount__ = None;
                let mut refunded_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RefundedAmount => {
                            if refunded_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("refundedAmount"));
                            }
                            refunded_amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RefundedAt => {
                            if refunded_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("refundedAt"));
                            }
                            refunded_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRefundEscrowResponse {
                    refunded_amount: refunded_amount__.unwrap_or_default(),
                    refunded_at: refunded_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgRefundEscrowResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgReleaseEscrow {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        if !self.amount.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgReleaseEscrow", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgReleaseEscrow {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "escrow_id",
            "escrowId",
            "amount",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            EscrowId,
            Amount,
            Reason,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            "amount" => Ok(GeneratedField::Amount),
                            "reason" => Ok(GeneratedField::Reason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgReleaseEscrow;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgReleaseEscrow")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgReleaseEscrow, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut escrow_id__ = None;
                let mut amount__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgReleaseEscrow {
                    sender: sender__.unwrap_or_default(),
                    escrow_id: escrow_id__.unwrap_or_default(),
                    amount: amount__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgReleaseEscrow", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgReleaseEscrowResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.released_amount.is_empty() {
            len += 1;
        }
        if self.released_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgReleaseEscrowResponse", len)?;
        if !self.released_amount.is_empty() {
            struct_ser.serialize_field("releasedAmount", &self.released_amount)?;
        }
        if self.released_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("releasedAt", ToString::to_string(&self.released_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgReleaseEscrowResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "released_amount",
            "releasedAmount",
            "released_at",
            "releasedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReleasedAmount,
            ReleasedAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "releasedAmount" | "released_amount" => Ok(GeneratedField::ReleasedAmount),
                            "releasedAt" | "released_at" => Ok(GeneratedField::ReleasedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgReleaseEscrowResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgReleaseEscrowResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgReleaseEscrowResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut released_amount__ = None;
                let mut released_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReleasedAmount => {
                            if released_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("releasedAmount"));
                            }
                            released_amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReleasedAt => {
                            if released_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("releasedAt"));
                            }
                            released_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgReleaseEscrowResponse {
                    released_amount: released_amount__.unwrap_or_default(),
                    released_at: released_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgReleaseEscrowResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSettleOrder {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if !self.usage_record_ids.is_empty() {
            len += 1;
        }
        if self.is_final {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgSettleOrder", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if !self.usage_record_ids.is_empty() {
            struct_ser.serialize_field("usageRecordIds", &self.usage_record_ids)?;
        }
        if self.is_final {
            struct_ser.serialize_field("isFinal", &self.is_final)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSettleOrder {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "order_id",
            "orderId",
            "usage_record_ids",
            "usageRecordIds",
            "is_final",
            "isFinal",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            OrderId,
            UsageRecordIds,
            IsFinal,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "sender" => Ok(GeneratedField::Sender),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "usageRecordIds" | "usage_record_ids" => Ok(GeneratedField::UsageRecordIds),
                            "isFinal" | "is_final" => Ok(GeneratedField::IsFinal),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSettleOrder;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgSettleOrder")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSettleOrder, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut order_id__ = None;
                let mut usage_record_ids__ = None;
                let mut is_final__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UsageRecordIds => {
                            if usage_record_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageRecordIds"));
                            }
                            usage_record_ids__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsFinal => {
                            if is_final__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isFinal"));
                            }
                            is_final__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSettleOrder {
                    sender: sender__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    usage_record_ids: usage_record_ids__.unwrap_or_default(),
                    is_final: is_final__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgSettleOrder", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSettleOrderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.settlement_id.is_empty() {
            len += 1;
        }
        if !self.total_amount.is_empty() {
            len += 1;
        }
        if !self.provider_share.is_empty() {
            len += 1;
        }
        if !self.platform_fee.is_empty() {
            len += 1;
        }
        if self.settled_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.settlement.v1.MsgSettleOrderResponse", len)?;
        if !self.settlement_id.is_empty() {
            struct_ser.serialize_field("settlementId", &self.settlement_id)?;
        }
        if !self.total_amount.is_empty() {
            struct_ser.serialize_field("totalAmount", &self.total_amount)?;
        }
        if !self.provider_share.is_empty() {
            struct_ser.serialize_field("providerShare", &self.provider_share)?;
        }
        if !self.platform_fee.is_empty() {
            struct_ser.serialize_field("platformFee", &self.platform_fee)?;
        }
        if self.settled_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("settledAt", ToString::to_string(&self.settled_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSettleOrderResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "settlement_id",
            "settlementId",
            "total_amount",
            "totalAmount",
            "provider_share",
            "providerShare",
            "platform_fee",
            "platformFee",
            "settled_at",
            "settledAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SettlementId,
            TotalAmount,
            ProviderShare,
            PlatformFee,
            SettledAt,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "settlementId" | "settlement_id" => Ok(GeneratedField::SettlementId),
                            "totalAmount" | "total_amount" => Ok(GeneratedField::TotalAmount),
                            "providerShare" | "provider_share" => Ok(GeneratedField::ProviderShare),
                            "platformFee" | "platform_fee" => Ok(GeneratedField::PlatformFee),
                            "settledAt" | "settled_at" => Ok(GeneratedField::SettledAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSettleOrderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.settlement.v1.MsgSettleOrderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSettleOrderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut settlement_id__ = None;
                let mut total_amount__ = None;
                let mut provider_share__ = None;
                let mut platform_fee__ = None;
                let mut settled_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SettlementId => {
                            if settlement_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("settlementId"));
                            }
                            settlement_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalAmount => {
                            if total_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalAmount"));
                            }
                            total_amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderShare => {
                            if provider_share__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerShare"));
                            }
                            provider_share__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PlatformFee => {
                            if platform_fee__.is_some() {
                                return Err(serde::de::Error::duplicate_field("platformFee"));
                            }
                            platform_fee__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SettledAt => {
                            if settled_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("settledAt"));
                            }
                            settled_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgSettleOrderResponse {
                    settlement_id: settlement_id__.unwrap_or_default(),
                    total_amount: total_amount__.unwrap_or_default(),
                    provider_share: provider_share__.unwrap_or_default(),
                    platform_fee: platform_fee__.unwrap_or_default(),
                    settled_at: settled_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.settlement.v1.MsgSettleOrderResponse", FIELDS, GeneratedVisitor)
    }
}

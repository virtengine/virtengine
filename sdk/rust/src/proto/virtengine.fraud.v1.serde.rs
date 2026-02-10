// @generated
impl serde::Serialize for AuditAction {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "AUDIT_ACTION_UNSPECIFIED",
            Self::Submitted => "AUDIT_ACTION_SUBMITTED",
            Self::Assigned => "AUDIT_ACTION_ASSIGNED",
            Self::StatusChanged => "AUDIT_ACTION_STATUS_CHANGED",
            Self::EvidenceViewed => "AUDIT_ACTION_EVIDENCE_VIEWED",
            Self::Resolved => "AUDIT_ACTION_RESOLVED",
            Self::Rejected => "AUDIT_ACTION_REJECTED",
            Self::Escalated => "AUDIT_ACTION_ESCALATED",
            Self::CommentAdded => "AUDIT_ACTION_COMMENT_ADDED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for AuditAction {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "AUDIT_ACTION_UNSPECIFIED",
            "AUDIT_ACTION_SUBMITTED",
            "AUDIT_ACTION_ASSIGNED",
            "AUDIT_ACTION_STATUS_CHANGED",
            "AUDIT_ACTION_EVIDENCE_VIEWED",
            "AUDIT_ACTION_RESOLVED",
            "AUDIT_ACTION_REJECTED",
            "AUDIT_ACTION_ESCALATED",
            "AUDIT_ACTION_COMMENT_ADDED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AuditAction;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                write!(formatter, "expected one of: {:?}", &FIELDS)
            }

            fn visit_i64<E>(self, v: i64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Signed(v), &self)
                    })
            }

            fn visit_u64<E>(self, v: u64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "AUDIT_ACTION_UNSPECIFIED" => Ok(AuditAction::Unspecified),
                    "AUDIT_ACTION_SUBMITTED" => Ok(AuditAction::Submitted),
                    "AUDIT_ACTION_ASSIGNED" => Ok(AuditAction::Assigned),
                    "AUDIT_ACTION_STATUS_CHANGED" => Ok(AuditAction::StatusChanged),
                    "AUDIT_ACTION_EVIDENCE_VIEWED" => Ok(AuditAction::EvidenceViewed),
                    "AUDIT_ACTION_RESOLVED" => Ok(AuditAction::Resolved),
                    "AUDIT_ACTION_REJECTED" => Ok(AuditAction::Rejected),
                    "AUDIT_ACTION_ESCALATED" => Ok(AuditAction::Escalated),
                    "AUDIT_ACTION_COMMENT_ADDED" => Ok(AuditAction::CommentAdded),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for EncryptedEvidence {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.algorithm_id.is_empty() {
            len += 1;
        }
        if !self.recipient_key_ids.is_empty() {
            len += 1;
        }
        if !self.encrypted_keys.is_empty() {
            len += 1;
        }
        if !self.nonce.is_empty() {
            len += 1;
        }
        if !self.ciphertext.is_empty() {
            len += 1;
        }
        if !self.sender_signature.is_empty() {
            len += 1;
        }
        if !self.sender_pub_key.is_empty() {
            len += 1;
        }
        if !self.content_type.is_empty() {
            len += 1;
        }
        if !self.evidence_hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.EncryptedEvidence", len)?;
        if !self.algorithm_id.is_empty() {
            struct_ser.serialize_field("algorithmId", &self.algorithm_id)?;
        }
        if !self.recipient_key_ids.is_empty() {
            struct_ser.serialize_field("recipientKeyIds", &self.recipient_key_ids)?;
        }
        if !self.encrypted_keys.is_empty() {
            struct_ser.serialize_field("encryptedKeys", &self.encrypted_keys.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.nonce.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nonce", pbjson::private::base64::encode(&self.nonce).as_str())?;
        }
        if !self.ciphertext.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("ciphertext", pbjson::private::base64::encode(&self.ciphertext).as_str())?;
        }
        if !self.sender_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("senderSignature", pbjson::private::base64::encode(&self.sender_signature).as_str())?;
        }
        if !self.sender_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("senderPubKey", pbjson::private::base64::encode(&self.sender_pub_key).as_str())?;
        }
        if !self.content_type.is_empty() {
            struct_ser.serialize_field("contentType", &self.content_type)?;
        }
        if !self.evidence_hash.is_empty() {
            struct_ser.serialize_field("evidenceHash", &self.evidence_hash)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EncryptedEvidence {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "algorithm_id",
            "algorithmId",
            "recipient_key_ids",
            "recipientKeyIds",
            "encrypted_keys",
            "encryptedKeys",
            "nonce",
            "ciphertext",
            "sender_signature",
            "senderSignature",
            "sender_pub_key",
            "senderPubKey",
            "content_type",
            "contentType",
            "evidence_hash",
            "evidenceHash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AlgorithmId,
            RecipientKeyIds,
            EncryptedKeys,
            Nonce,
            Ciphertext,
            SenderSignature,
            SenderPubKey,
            ContentType,
            EvidenceHash,
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
                            "algorithmId" | "algorithm_id" => Ok(GeneratedField::AlgorithmId),
                            "recipientKeyIds" | "recipient_key_ids" => Ok(GeneratedField::RecipientKeyIds),
                            "encryptedKeys" | "encrypted_keys" => Ok(GeneratedField::EncryptedKeys),
                            "nonce" => Ok(GeneratedField::Nonce),
                            "ciphertext" => Ok(GeneratedField::Ciphertext),
                            "senderSignature" | "sender_signature" => Ok(GeneratedField::SenderSignature),
                            "senderPubKey" | "sender_pub_key" => Ok(GeneratedField::SenderPubKey),
                            "contentType" | "content_type" => Ok(GeneratedField::ContentType),
                            "evidenceHash" | "evidence_hash" => Ok(GeneratedField::EvidenceHash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EncryptedEvidence;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.EncryptedEvidence")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EncryptedEvidence, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut algorithm_id__ = None;
                let mut recipient_key_ids__ = None;
                let mut encrypted_keys__ = None;
                let mut nonce__ = None;
                let mut ciphertext__ = None;
                let mut sender_signature__ = None;
                let mut sender_pub_key__ = None;
                let mut content_type__ = None;
                let mut evidence_hash__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AlgorithmId => {
                            if algorithm_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmId"));
                            }
                            algorithm_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecipientKeyIds => {
                            if recipient_key_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientKeyIds"));
                            }
                            recipient_key_ids__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptedKeys => {
                            if encrypted_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedKeys"));
                            }
                            encrypted_keys__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::Nonce => {
                            if nonce__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nonce"));
                            }
                            nonce__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Ciphertext => {
                            if ciphertext__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ciphertext"));
                            }
                            ciphertext__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SenderSignature => {
                            if sender_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("senderSignature"));
                            }
                            sender_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SenderPubKey => {
                            if sender_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("senderPubKey"));
                            }
                            sender_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ContentType => {
                            if content_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("contentType"));
                            }
                            content_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EvidenceHash => {
                            if evidence_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidenceHash"));
                            }
                            evidence_hash__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EncryptedEvidence {
                    algorithm_id: algorithm_id__.unwrap_or_default(),
                    recipient_key_ids: recipient_key_ids__.unwrap_or_default(),
                    encrypted_keys: encrypted_keys__.unwrap_or_default(),
                    nonce: nonce__.unwrap_or_default(),
                    ciphertext: ciphertext__.unwrap_or_default(),
                    sender_signature: sender_signature__.unwrap_or_default(),
                    sender_pub_key: sender_pub_key__.unwrap_or_default(),
                    content_type: content_type__.unwrap_or_default(),
                    evidence_hash: evidence_hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.EncryptedEvidence", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FraudAuditLog {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.id.is_empty() {
            len += 1;
        }
        if !self.report_id.is_empty() {
            len += 1;
        }
        if self.action != 0 {
            len += 1;
        }
        if !self.actor.is_empty() {
            len += 1;
        }
        if self.previous_status != 0 {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if !self.details.is_empty() {
            len += 1;
        }
        if self.timestamp.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if !self.tx_hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.FraudAuditLog", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if self.action != 0 {
            let v = AuditAction::try_from(self.action)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.action)))?;
            struct_ser.serialize_field("action", &v)?;
        }
        if !self.actor.is_empty() {
            struct_ser.serialize_field("actor", &self.actor)?;
        }
        if self.previous_status != 0 {
            let v = FraudReportStatus::try_from(self.previous_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_status)))?;
            struct_ser.serialize_field("previousStatus", &v)?;
        }
        if self.new_status != 0 {
            let v = FraudReportStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if !self.details.is_empty() {
            struct_ser.serialize_field("details", &self.details)?;
        }
        if let Some(v) = self.timestamp.as_ref() {
            struct_ser.serialize_field("timestamp", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if !self.tx_hash.is_empty() {
            struct_ser.serialize_field("txHash", &self.tx_hash)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for FraudAuditLog {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "report_id",
            "reportId",
            "action",
            "actor",
            "previous_status",
            "previousStatus",
            "new_status",
            "newStatus",
            "details",
            "timestamp",
            "block_height",
            "blockHeight",
            "tx_hash",
            "txHash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            ReportId,
            Action,
            Actor,
            PreviousStatus,
            NewStatus,
            Details,
            Timestamp,
            BlockHeight,
            TxHash,
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
                            "id" => Ok(GeneratedField::Id),
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "action" => Ok(GeneratedField::Action),
                            "actor" => Ok(GeneratedField::Actor),
                            "previousStatus" | "previous_status" => Ok(GeneratedField::PreviousStatus),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
                            "details" => Ok(GeneratedField::Details),
                            "timestamp" => Ok(GeneratedField::Timestamp),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "txHash" | "tx_hash" => Ok(GeneratedField::TxHash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FraudAuditLog;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.FraudAuditLog")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<FraudAuditLog, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut report_id__ = None;
                let mut action__ = None;
                let mut actor__ = None;
                let mut previous_status__ = None;
                let mut new_status__ = None;
                let mut details__ = None;
                let mut timestamp__ = None;
                let mut block_height__ = None;
                let mut tx_hash__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Action => {
                            if action__.is_some() {
                                return Err(serde::de::Error::duplicate_field("action"));
                            }
                            action__ = Some(map_.next_value::<AuditAction>()? as i32);
                        }
                        GeneratedField::Actor => {
                            if actor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("actor"));
                            }
                            actor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PreviousStatus => {
                            if previous_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousStatus"));
                            }
                            previous_status__ = Some(map_.next_value::<FraudReportStatus>()? as i32);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<FraudReportStatus>()? as i32);
                        }
                        GeneratedField::Details => {
                            if details__.is_some() {
                                return Err(serde::de::Error::duplicate_field("details"));
                            }
                            details__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TxHash => {
                            if tx_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("txHash"));
                            }
                            tx_hash__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(FraudAuditLog {
                    id: id__.unwrap_or_default(),
                    report_id: report_id__.unwrap_or_default(),
                    action: action__.unwrap_or_default(),
                    actor: actor__.unwrap_or_default(),
                    previous_status: previous_status__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    details: details__.unwrap_or_default(),
                    timestamp: timestamp__,
                    block_height: block_height__.unwrap_or_default(),
                    tx_hash: tx_hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.FraudAuditLog", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FraudCategory {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "FRAUD_CATEGORY_UNSPECIFIED",
            Self::FakeIdentity => "FRAUD_CATEGORY_FAKE_IDENTITY",
            Self::PaymentFraud => "FRAUD_CATEGORY_PAYMENT_FRAUD",
            Self::ServiceMisrepresentation => "FRAUD_CATEGORY_SERVICE_MISREPRESENTATION",
            Self::ResourceAbuse => "FRAUD_CATEGORY_RESOURCE_ABUSE",
            Self::SybilAttack => "FRAUD_CATEGORY_SYBIL_ATTACK",
            Self::MaliciousContent => "FRAUD_CATEGORY_MALICIOUS_CONTENT",
            Self::TermsViolation => "FRAUD_CATEGORY_TERMS_VIOLATION",
            Self::Other => "FRAUD_CATEGORY_OTHER",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for FraudCategory {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "FRAUD_CATEGORY_UNSPECIFIED",
            "FRAUD_CATEGORY_FAKE_IDENTITY",
            "FRAUD_CATEGORY_PAYMENT_FRAUD",
            "FRAUD_CATEGORY_SERVICE_MISREPRESENTATION",
            "FRAUD_CATEGORY_RESOURCE_ABUSE",
            "FRAUD_CATEGORY_SYBIL_ATTACK",
            "FRAUD_CATEGORY_MALICIOUS_CONTENT",
            "FRAUD_CATEGORY_TERMS_VIOLATION",
            "FRAUD_CATEGORY_OTHER",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FraudCategory;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                write!(formatter, "expected one of: {:?}", &FIELDS)
            }

            fn visit_i64<E>(self, v: i64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Signed(v), &self)
                    })
            }

            fn visit_u64<E>(self, v: u64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "FRAUD_CATEGORY_UNSPECIFIED" => Ok(FraudCategory::Unspecified),
                    "FRAUD_CATEGORY_FAKE_IDENTITY" => Ok(FraudCategory::FakeIdentity),
                    "FRAUD_CATEGORY_PAYMENT_FRAUD" => Ok(FraudCategory::PaymentFraud),
                    "FRAUD_CATEGORY_SERVICE_MISREPRESENTATION" => Ok(FraudCategory::ServiceMisrepresentation),
                    "FRAUD_CATEGORY_RESOURCE_ABUSE" => Ok(FraudCategory::ResourceAbuse),
                    "FRAUD_CATEGORY_SYBIL_ATTACK" => Ok(FraudCategory::SybilAttack),
                    "FRAUD_CATEGORY_MALICIOUS_CONTENT" => Ok(FraudCategory::MaliciousContent),
                    "FRAUD_CATEGORY_TERMS_VIOLATION" => Ok(FraudCategory::TermsViolation),
                    "FRAUD_CATEGORY_OTHER" => Ok(FraudCategory::Other),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for FraudReport {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.id.is_empty() {
            len += 1;
        }
        if !self.reporter.is_empty() {
            len += 1;
        }
        if !self.reported_party.is_empty() {
            len += 1;
        }
        if self.category != 0 {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.assigned_moderator.is_empty() {
            len += 1;
        }
        if self.resolution != 0 {
            len += 1;
        }
        if !self.resolution_notes.is_empty() {
            len += 1;
        }
        if self.submitted_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.resolved_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if !self.content_hash.is_empty() {
            len += 1;
        }
        if !self.related_order_ids.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.FraudReport", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", &self.reporter)?;
        }
        if !self.reported_party.is_empty() {
            struct_ser.serialize_field("reportedParty", &self.reported_party)?;
        }
        if self.category != 0 {
            let v = FraudCategory::try_from(self.category)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.category)))?;
            struct_ser.serialize_field("category", &v)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        if self.status != 0 {
            let v = FraudReportStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.assigned_moderator.is_empty() {
            struct_ser.serialize_field("assignedModerator", &self.assigned_moderator)?;
        }
        if self.resolution != 0 {
            let v = ResolutionType::try_from(self.resolution)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.resolution)))?;
            struct_ser.serialize_field("resolution", &v)?;
        }
        if !self.resolution_notes.is_empty() {
            struct_ser.serialize_field("resolutionNotes", &self.resolution_notes)?;
        }
        if let Some(v) = self.submitted_at.as_ref() {
            struct_ser.serialize_field("submittedAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if let Some(v) = self.resolved_at.as_ref() {
            struct_ser.serialize_field("resolvedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if !self.content_hash.is_empty() {
            struct_ser.serialize_field("contentHash", &self.content_hash)?;
        }
        if !self.related_order_ids.is_empty() {
            struct_ser.serialize_field("relatedOrderIds", &self.related_order_ids)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for FraudReport {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "reporter",
            "reported_party",
            "reportedParty",
            "category",
            "description",
            "evidence",
            "status",
            "assigned_moderator",
            "assignedModerator",
            "resolution",
            "resolution_notes",
            "resolutionNotes",
            "submitted_at",
            "submittedAt",
            "updated_at",
            "updatedAt",
            "resolved_at",
            "resolvedAt",
            "block_height",
            "blockHeight",
            "content_hash",
            "contentHash",
            "related_order_ids",
            "relatedOrderIds",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Reporter,
            ReportedParty,
            Category,
            Description,
            Evidence,
            Status,
            AssignedModerator,
            Resolution,
            ResolutionNotes,
            SubmittedAt,
            UpdatedAt,
            ResolvedAt,
            BlockHeight,
            ContentHash,
            RelatedOrderIds,
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
                            "id" => Ok(GeneratedField::Id),
                            "reporter" => Ok(GeneratedField::Reporter),
                            "reportedParty" | "reported_party" => Ok(GeneratedField::ReportedParty),
                            "category" => Ok(GeneratedField::Category),
                            "description" => Ok(GeneratedField::Description),
                            "evidence" => Ok(GeneratedField::Evidence),
                            "status" => Ok(GeneratedField::Status),
                            "assignedModerator" | "assigned_moderator" => Ok(GeneratedField::AssignedModerator),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "resolutionNotes" | "resolution_notes" => Ok(GeneratedField::ResolutionNotes),
                            "submittedAt" | "submitted_at" => Ok(GeneratedField::SubmittedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "resolvedAt" | "resolved_at" => Ok(GeneratedField::ResolvedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "contentHash" | "content_hash" => Ok(GeneratedField::ContentHash),
                            "relatedOrderIds" | "related_order_ids" => Ok(GeneratedField::RelatedOrderIds),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FraudReport;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.FraudReport")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<FraudReport, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut reporter__ = None;
                let mut reported_party__ = None;
                let mut category__ = None;
                let mut description__ = None;
                let mut evidence__ = None;
                let mut status__ = None;
                let mut assigned_moderator__ = None;
                let mut resolution__ = None;
                let mut resolution_notes__ = None;
                let mut submitted_at__ = None;
                let mut updated_at__ = None;
                let mut resolved_at__ = None;
                let mut block_height__ = None;
                let mut content_hash__ = None;
                let mut related_order_ids__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportedParty => {
                            if reported_party__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportedParty"));
                            }
                            reported_party__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value::<FraudCategory>()? as i32);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<FraudReportStatus>()? as i32);
                        }
                        GeneratedField::AssignedModerator => {
                            if assigned_moderator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedModerator"));
                            }
                            assigned_moderator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value::<ResolutionType>()? as i32);
                        }
                        GeneratedField::ResolutionNotes => {
                            if resolution_notes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolutionNotes"));
                            }
                            resolution_notes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubmittedAt => {
                            if submitted_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submittedAt"));
                            }
                            submitted_at__ = map_.next_value()?;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                        GeneratedField::ResolvedAt => {
                            if resolved_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolvedAt"));
                            }
                            resolved_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ContentHash => {
                            if content_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("contentHash"));
                            }
                            content_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RelatedOrderIds => {
                            if related_order_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("relatedOrderIds"));
                            }
                            related_order_ids__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(FraudReport {
                    id: id__.unwrap_or_default(),
                    reporter: reporter__.unwrap_or_default(),
                    reported_party: reported_party__.unwrap_or_default(),
                    category: category__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    assigned_moderator: assigned_moderator__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    resolution_notes: resolution_notes__.unwrap_or_default(),
                    submitted_at: submitted_at__,
                    updated_at: updated_at__,
                    resolved_at: resolved_at__,
                    block_height: block_height__.unwrap_or_default(),
                    content_hash: content_hash__.unwrap_or_default(),
                    related_order_ids: related_order_ids__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.FraudReport", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FraudReportStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "FRAUD_REPORT_STATUS_UNSPECIFIED",
            Self::Submitted => "FRAUD_REPORT_STATUS_SUBMITTED",
            Self::Reviewing => "FRAUD_REPORT_STATUS_REVIEWING",
            Self::Resolved => "FRAUD_REPORT_STATUS_RESOLVED",
            Self::Rejected => "FRAUD_REPORT_STATUS_REJECTED",
            Self::Escalated => "FRAUD_REPORT_STATUS_ESCALATED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for FraudReportStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "FRAUD_REPORT_STATUS_UNSPECIFIED",
            "FRAUD_REPORT_STATUS_SUBMITTED",
            "FRAUD_REPORT_STATUS_REVIEWING",
            "FRAUD_REPORT_STATUS_RESOLVED",
            "FRAUD_REPORT_STATUS_REJECTED",
            "FRAUD_REPORT_STATUS_ESCALATED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FraudReportStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                write!(formatter, "expected one of: {:?}", &FIELDS)
            }

            fn visit_i64<E>(self, v: i64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Signed(v), &self)
                    })
            }

            fn visit_u64<E>(self, v: u64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "FRAUD_REPORT_STATUS_UNSPECIFIED" => Ok(FraudReportStatus::Unspecified),
                    "FRAUD_REPORT_STATUS_SUBMITTED" => Ok(FraudReportStatus::Submitted),
                    "FRAUD_REPORT_STATUS_REVIEWING" => Ok(FraudReportStatus::Reviewing),
                    "FRAUD_REPORT_STATUS_RESOLVED" => Ok(FraudReportStatus::Resolved),
                    "FRAUD_REPORT_STATUS_REJECTED" => Ok(FraudReportStatus::Rejected),
                    "FRAUD_REPORT_STATUS_ESCALATED" => Ok(FraudReportStatus::Escalated),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for GenesisState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.params.is_some() {
            len += 1;
        }
        if !self.reports.is_empty() {
            len += 1;
        }
        if !self.audit_logs.is_empty() {
            len += 1;
        }
        if !self.moderator_queue.is_empty() {
            len += 1;
        }
        if self.next_fraud_report_sequence != 0 {
            len += 1;
        }
        if self.next_audit_log_sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.reports.is_empty() {
            struct_ser.serialize_field("reports", &self.reports)?;
        }
        if !self.audit_logs.is_empty() {
            struct_ser.serialize_field("auditLogs", &self.audit_logs)?;
        }
        if !self.moderator_queue.is_empty() {
            struct_ser.serialize_field("moderatorQueue", &self.moderator_queue)?;
        }
        if self.next_fraud_report_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nextFraudReportSequence", ToString::to_string(&self.next_fraud_report_sequence).as_str())?;
        }
        if self.next_audit_log_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nextAuditLogSequence", ToString::to_string(&self.next_audit_log_sequence).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GenesisState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "params",
            "reports",
            "audit_logs",
            "auditLogs",
            "moderator_queue",
            "moderatorQueue",
            "next_fraud_report_sequence",
            "nextFraudReportSequence",
            "next_audit_log_sequence",
            "nextAuditLogSequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            Reports,
            AuditLogs,
            ModeratorQueue,
            NextFraudReportSequence,
            NextAuditLogSequence,
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
                            "params" => Ok(GeneratedField::Params),
                            "reports" => Ok(GeneratedField::Reports),
                            "auditLogs" | "audit_logs" => Ok(GeneratedField::AuditLogs),
                            "moderatorQueue" | "moderator_queue" => Ok(GeneratedField::ModeratorQueue),
                            "nextFraudReportSequence" | "next_fraud_report_sequence" => Ok(GeneratedField::NextFraudReportSequence),
                            "nextAuditLogSequence" | "next_audit_log_sequence" => Ok(GeneratedField::NextAuditLogSequence),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GenesisState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut reports__ = None;
                let mut audit_logs__ = None;
                let mut moderator_queue__ = None;
                let mut next_fraud_report_sequence__ = None;
                let mut next_audit_log_sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Reports => {
                            if reports__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reports"));
                            }
                            reports__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AuditLogs => {
                            if audit_logs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditLogs"));
                            }
                            audit_logs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModeratorQueue => {
                            if moderator_queue__.is_some() {
                                return Err(serde::de::Error::duplicate_field("moderatorQueue"));
                            }
                            moderator_queue__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NextFraudReportSequence => {
                            if next_fraud_report_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nextFraudReportSequence"));
                            }
                            next_fraud_report_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NextAuditLogSequence => {
                            if next_audit_log_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nextAuditLogSequence"));
                            }
                            next_audit_log_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    reports: reports__.unwrap_or_default(),
                    audit_logs: audit_logs__.unwrap_or_default(),
                    moderator_queue: moderator_queue__.unwrap_or_default(),
                    next_fraud_report_sequence: next_fraud_report_sequence__.unwrap_or_default(),
                    next_audit_log_sequence: next_audit_log_sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ModeratorQueueEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.report_id.is_empty() {
            len += 1;
        }
        if self.priority != 0 {
            len += 1;
        }
        if self.queued_at.is_some() {
            len += 1;
        }
        if self.category != 0 {
            len += 1;
        }
        if !self.assigned_to.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.ModeratorQueueEntry", len)?;
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if self.priority != 0 {
            struct_ser.serialize_field("priority", &self.priority)?;
        }
        if let Some(v) = self.queued_at.as_ref() {
            struct_ser.serialize_field("queuedAt", v)?;
        }
        if self.category != 0 {
            let v = FraudCategory::try_from(self.category)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.category)))?;
            struct_ser.serialize_field("category", &v)?;
        }
        if !self.assigned_to.is_empty() {
            struct_ser.serialize_field("assignedTo", &self.assigned_to)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ModeratorQueueEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "report_id",
            "reportId",
            "priority",
            "queued_at",
            "queuedAt",
            "category",
            "assigned_to",
            "assignedTo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReportId,
            Priority,
            QueuedAt,
            Category,
            AssignedTo,
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
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "priority" => Ok(GeneratedField::Priority),
                            "queuedAt" | "queued_at" => Ok(GeneratedField::QueuedAt),
                            "category" => Ok(GeneratedField::Category),
                            "assignedTo" | "assigned_to" => Ok(GeneratedField::AssignedTo),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModeratorQueueEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.ModeratorQueueEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ModeratorQueueEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut report_id__ = None;
                let mut priority__ = None;
                let mut queued_at__ = None;
                let mut category__ = None;
                let mut assigned_to__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Priority => {
                            if priority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("priority"));
                            }
                            priority__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::QueuedAt => {
                            if queued_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queuedAt"));
                            }
                            queued_at__ = map_.next_value()?;
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value::<FraudCategory>()? as i32);
                        }
                        GeneratedField::AssignedTo => {
                            if assigned_to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedTo"));
                            }
                            assigned_to__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ModeratorQueueEntry {
                    report_id: report_id__.unwrap_or_default(),
                    priority: priority__.unwrap_or_default(),
                    queued_at: queued_at__,
                    category: category__.unwrap_or_default(),
                    assigned_to: assigned_to__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.ModeratorQueueEntry", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAssignModerator {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.moderator.is_empty() {
            len += 1;
        }
        if !self.report_id.is_empty() {
            len += 1;
        }
        if !self.assign_to.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgAssignModerator", len)?;
        if !self.moderator.is_empty() {
            struct_ser.serialize_field("moderator", &self.moderator)?;
        }
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if !self.assign_to.is_empty() {
            struct_ser.serialize_field("assignTo", &self.assign_to)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAssignModerator {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "moderator",
            "report_id",
            "reportId",
            "assign_to",
            "assignTo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Moderator,
            ReportId,
            AssignTo,
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
                            "moderator" => Ok(GeneratedField::Moderator),
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "assignTo" | "assign_to" => Ok(GeneratedField::AssignTo),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAssignModerator;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgAssignModerator")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAssignModerator, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut moderator__ = None;
                let mut report_id__ = None;
                let mut assign_to__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Moderator => {
                            if moderator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("moderator"));
                            }
                            moderator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AssignTo => {
                            if assign_to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignTo"));
                            }
                            assign_to__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgAssignModerator {
                    moderator: moderator__.unwrap_or_default(),
                    report_id: report_id__.unwrap_or_default(),
                    assign_to: assign_to__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgAssignModerator", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAssignModeratorResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgAssignModeratorResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAssignModeratorResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAssignModeratorResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgAssignModeratorResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAssignModeratorResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgAssignModeratorResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgAssignModeratorResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgEscalateFraudReport {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.moderator.is_empty() {
            len += 1;
        }
        if !self.report_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgEscalateFraudReport", len)?;
        if !self.moderator.is_empty() {
            struct_ser.serialize_field("moderator", &self.moderator)?;
        }
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgEscalateFraudReport {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "moderator",
            "report_id",
            "reportId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Moderator,
            ReportId,
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
                            "moderator" => Ok(GeneratedField::Moderator),
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
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
            type Value = MsgEscalateFraudReport;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgEscalateFraudReport")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgEscalateFraudReport, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut moderator__ = None;
                let mut report_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Moderator => {
                            if moderator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("moderator"));
                            }
                            moderator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgEscalateFraudReport {
                    moderator: moderator__.unwrap_or_default(),
                    report_id: report_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgEscalateFraudReport", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgEscalateFraudReportResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgEscalateFraudReportResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgEscalateFraudReportResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgEscalateFraudReportResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgEscalateFraudReportResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgEscalateFraudReportResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgEscalateFraudReportResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgEscalateFraudReportResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRejectFraudReport {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.moderator.is_empty() {
            len += 1;
        }
        if !self.report_id.is_empty() {
            len += 1;
        }
        if !self.notes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgRejectFraudReport", len)?;
        if !self.moderator.is_empty() {
            struct_ser.serialize_field("moderator", &self.moderator)?;
        }
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if !self.notes.is_empty() {
            struct_ser.serialize_field("notes", &self.notes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRejectFraudReport {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "moderator",
            "report_id",
            "reportId",
            "notes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Moderator,
            ReportId,
            Notes,
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
                            "moderator" => Ok(GeneratedField::Moderator),
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "notes" => Ok(GeneratedField::Notes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRejectFraudReport;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgRejectFraudReport")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRejectFraudReport, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut moderator__ = None;
                let mut report_id__ = None;
                let mut notes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Moderator => {
                            if moderator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("moderator"));
                            }
                            moderator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Notes => {
                            if notes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("notes"));
                            }
                            notes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRejectFraudReport {
                    moderator: moderator__.unwrap_or_default(),
                    report_id: report_id__.unwrap_or_default(),
                    notes: notes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgRejectFraudReport", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRejectFraudReportResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgRejectFraudReportResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRejectFraudReportResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRejectFraudReportResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgRejectFraudReportResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRejectFraudReportResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRejectFraudReportResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgRejectFraudReportResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveFraudReport {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.moderator.is_empty() {
            len += 1;
        }
        if !self.report_id.is_empty() {
            len += 1;
        }
        if self.resolution != 0 {
            len += 1;
        }
        if !self.notes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgResolveFraudReport", len)?;
        if !self.moderator.is_empty() {
            struct_ser.serialize_field("moderator", &self.moderator)?;
        }
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if self.resolution != 0 {
            let v = ResolutionType::try_from(self.resolution)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.resolution)))?;
            struct_ser.serialize_field("resolution", &v)?;
        }
        if !self.notes.is_empty() {
            struct_ser.serialize_field("notes", &self.notes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveFraudReport {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "moderator",
            "report_id",
            "reportId",
            "resolution",
            "notes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Moderator,
            ReportId,
            Resolution,
            Notes,
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
                            "moderator" => Ok(GeneratedField::Moderator),
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "notes" => Ok(GeneratedField::Notes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgResolveFraudReport;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgResolveFraudReport")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveFraudReport, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut moderator__ = None;
                let mut report_id__ = None;
                let mut resolution__ = None;
                let mut notes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Moderator => {
                            if moderator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("moderator"));
                            }
                            moderator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value::<ResolutionType>()? as i32);
                        }
                        GeneratedField::Notes => {
                            if notes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("notes"));
                            }
                            notes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgResolveFraudReport {
                    moderator: moderator__.unwrap_or_default(),
                    report_id: report_id__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    notes: notes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgResolveFraudReport", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveFraudReportResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgResolveFraudReportResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveFraudReportResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgResolveFraudReportResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgResolveFraudReportResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveFraudReportResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgResolveFraudReportResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgResolveFraudReportResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitFraudReport {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reporter.is_empty() {
            len += 1;
        }
        if !self.reported_party.is_empty() {
            len += 1;
        }
        if self.category != 0 {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        if !self.related_order_ids.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgSubmitFraudReport", len)?;
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", &self.reporter)?;
        }
        if !self.reported_party.is_empty() {
            struct_ser.serialize_field("reportedParty", &self.reported_party)?;
        }
        if self.category != 0 {
            let v = FraudCategory::try_from(self.category)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.category)))?;
            struct_ser.serialize_field("category", &v)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        if !self.related_order_ids.is_empty() {
            struct_ser.serialize_field("relatedOrderIds", &self.related_order_ids)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitFraudReport {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reporter",
            "reported_party",
            "reportedParty",
            "category",
            "description",
            "evidence",
            "related_order_ids",
            "relatedOrderIds",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reporter,
            ReportedParty,
            Category,
            Description,
            Evidence,
            RelatedOrderIds,
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
                            "reporter" => Ok(GeneratedField::Reporter),
                            "reportedParty" | "reported_party" => Ok(GeneratedField::ReportedParty),
                            "category" => Ok(GeneratedField::Category),
                            "description" => Ok(GeneratedField::Description),
                            "evidence" => Ok(GeneratedField::Evidence),
                            "relatedOrderIds" | "related_order_ids" => Ok(GeneratedField::RelatedOrderIds),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitFraudReport;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgSubmitFraudReport")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitFraudReport, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reporter__ = None;
                let mut reported_party__ = None;
                let mut category__ = None;
                let mut description__ = None;
                let mut evidence__ = None;
                let mut related_order_ids__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportedParty => {
                            if reported_party__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportedParty"));
                            }
                            reported_party__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value::<FraudCategory>()? as i32);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RelatedOrderIds => {
                            if related_order_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("relatedOrderIds"));
                            }
                            related_order_ids__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitFraudReport {
                    reporter: reporter__.unwrap_or_default(),
                    reported_party: reported_party__.unwrap_or_default(),
                    category: category__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                    related_order_ids: related_order_ids__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgSubmitFraudReport", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitFraudReportResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.report_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgSubmitFraudReportResponse", len)?;
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitFraudReportResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "report_id",
            "reportId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReportId,
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
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitFraudReportResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgSubmitFraudReportResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitFraudReportResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut report_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitFraudReportResponse {
                    report_id: report_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgSubmitFraudReportResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.authority.is_empty() {
            len += 1;
        }
        if self.params.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgUpdateParams", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "params",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            Params,
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
                            "authority" => Ok(GeneratedField::Authority),
                            "params" => Ok(GeneratedField::Params),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgUpdateParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut params__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgUpdateParams {
                    authority: authority__.unwrap_or_default(),
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateParamsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgUpdateParamsResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateParamsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgUpdateParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateParamsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateParamsResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateReportStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.moderator.is_empty() {
            len += 1;
        }
        if !self.report_id.is_empty() {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if !self.notes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgUpdateReportStatus", len)?;
        if !self.moderator.is_empty() {
            struct_ser.serialize_field("moderator", &self.moderator)?;
        }
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if self.new_status != 0 {
            let v = FraudReportStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if !self.notes.is_empty() {
            struct_ser.serialize_field("notes", &self.notes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateReportStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "moderator",
            "report_id",
            "reportId",
            "new_status",
            "newStatus",
            "notes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Moderator,
            ReportId,
            NewStatus,
            Notes,
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
                            "moderator" => Ok(GeneratedField::Moderator),
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
                            "notes" => Ok(GeneratedField::Notes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateReportStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgUpdateReportStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateReportStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut moderator__ = None;
                let mut report_id__ = None;
                let mut new_status__ = None;
                let mut notes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Moderator => {
                            if moderator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("moderator"));
                            }
                            moderator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<FraudReportStatus>()? as i32);
                        }
                        GeneratedField::Notes => {
                            if notes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("notes"));
                            }
                            notes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateReportStatus {
                    moderator: moderator__.unwrap_or_default(),
                    report_id: report_id__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    notes: notes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgUpdateReportStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateReportStatusResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.MsgUpdateReportStatusResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateReportStatusResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateReportStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.MsgUpdateReportStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateReportStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateReportStatusResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.MsgUpdateReportStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Params {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.min_description_length != 0 {
            len += 1;
        }
        if self.max_description_length != 0 {
            len += 1;
        }
        if self.max_evidence_count != 0 {
            len += 1;
        }
        if self.max_evidence_size_bytes != 0 {
            len += 1;
        }
        if self.auto_assign_enabled {
            len += 1;
        }
        if self.escalation_threshold_days != 0 {
            len += 1;
        }
        if self.report_retention_days != 0 {
            len += 1;
        }
        if self.audit_log_retention_days != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.Params", len)?;
        if self.min_description_length != 0 {
            struct_ser.serialize_field("minDescriptionLength", &self.min_description_length)?;
        }
        if self.max_description_length != 0 {
            struct_ser.serialize_field("maxDescriptionLength", &self.max_description_length)?;
        }
        if self.max_evidence_count != 0 {
            struct_ser.serialize_field("maxEvidenceCount", &self.max_evidence_count)?;
        }
        if self.max_evidence_size_bytes != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxEvidenceSizeBytes", ToString::to_string(&self.max_evidence_size_bytes).as_str())?;
        }
        if self.auto_assign_enabled {
            struct_ser.serialize_field("autoAssignEnabled", &self.auto_assign_enabled)?;
        }
        if self.escalation_threshold_days != 0 {
            struct_ser.serialize_field("escalationThresholdDays", &self.escalation_threshold_days)?;
        }
        if self.report_retention_days != 0 {
            struct_ser.serialize_field("reportRetentionDays", &self.report_retention_days)?;
        }
        if self.audit_log_retention_days != 0 {
            struct_ser.serialize_field("auditLogRetentionDays", &self.audit_log_retention_days)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Params {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "min_description_length",
            "minDescriptionLength",
            "max_description_length",
            "maxDescriptionLength",
            "max_evidence_count",
            "maxEvidenceCount",
            "max_evidence_size_bytes",
            "maxEvidenceSizeBytes",
            "auto_assign_enabled",
            "autoAssignEnabled",
            "escalation_threshold_days",
            "escalationThresholdDays",
            "report_retention_days",
            "reportRetentionDays",
            "audit_log_retention_days",
            "auditLogRetentionDays",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MinDescriptionLength,
            MaxDescriptionLength,
            MaxEvidenceCount,
            MaxEvidenceSizeBytes,
            AutoAssignEnabled,
            EscalationThresholdDays,
            ReportRetentionDays,
            AuditLogRetentionDays,
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
                            "minDescriptionLength" | "min_description_length" => Ok(GeneratedField::MinDescriptionLength),
                            "maxDescriptionLength" | "max_description_length" => Ok(GeneratedField::MaxDescriptionLength),
                            "maxEvidenceCount" | "max_evidence_count" => Ok(GeneratedField::MaxEvidenceCount),
                            "maxEvidenceSizeBytes" | "max_evidence_size_bytes" => Ok(GeneratedField::MaxEvidenceSizeBytes),
                            "autoAssignEnabled" | "auto_assign_enabled" => Ok(GeneratedField::AutoAssignEnabled),
                            "escalationThresholdDays" | "escalation_threshold_days" => Ok(GeneratedField::EscalationThresholdDays),
                            "reportRetentionDays" | "report_retention_days" => Ok(GeneratedField::ReportRetentionDays),
                            "auditLogRetentionDays" | "audit_log_retention_days" => Ok(GeneratedField::AuditLogRetentionDays),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Params;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut min_description_length__ = None;
                let mut max_description_length__ = None;
                let mut max_evidence_count__ = None;
                let mut max_evidence_size_bytes__ = None;
                let mut auto_assign_enabled__ = None;
                let mut escalation_threshold_days__ = None;
                let mut report_retention_days__ = None;
                let mut audit_log_retention_days__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MinDescriptionLength => {
                            if min_description_length__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minDescriptionLength"));
                            }
                            min_description_length__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxDescriptionLength => {
                            if max_description_length__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxDescriptionLength"));
                            }
                            max_description_length__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxEvidenceCount => {
                            if max_evidence_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxEvidenceCount"));
                            }
                            max_evidence_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxEvidenceSizeBytes => {
                            if max_evidence_size_bytes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxEvidenceSizeBytes"));
                            }
                            max_evidence_size_bytes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AutoAssignEnabled => {
                            if auto_assign_enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("autoAssignEnabled"));
                            }
                            auto_assign_enabled__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscalationThresholdDays => {
                            if escalation_threshold_days__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escalationThresholdDays"));
                            }
                            escalation_threshold_days__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReportRetentionDays => {
                            if report_retention_days__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportRetentionDays"));
                            }
                            report_retention_days__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AuditLogRetentionDays => {
                            if audit_log_retention_days__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditLogRetentionDays"));
                            }
                            audit_log_retention_days__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Params {
                    min_description_length: min_description_length__.unwrap_or_default(),
                    max_description_length: max_description_length__.unwrap_or_default(),
                    max_evidence_count: max_evidence_count__.unwrap_or_default(),
                    max_evidence_size_bytes: max_evidence_size_bytes__.unwrap_or_default(),
                    auto_assign_enabled: auto_assign_enabled__.unwrap_or_default(),
                    escalation_threshold_days: escalation_threshold_days__.unwrap_or_default(),
                    report_retention_days: report_retention_days__.unwrap_or_default(),
                    audit_log_retention_days: audit_log_retention_days__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAuditLogRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.report_id.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryAuditLogRequest", len)?;
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAuditLogRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "report_id",
            "reportId",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReportId,
            Pagination,
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
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAuditLogRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryAuditLogRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAuditLogRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut report_id__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAuditLogRequest {
                    report_id: report_id__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryAuditLogRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAuditLogResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.audit_logs.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryAuditLogResponse", len)?;
        if !self.audit_logs.is_empty() {
            struct_ser.serialize_field("auditLogs", &self.audit_logs)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAuditLogResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "audit_logs",
            "auditLogs",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AuditLogs,
            Pagination,
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
                            "auditLogs" | "audit_logs" => Ok(GeneratedField::AuditLogs),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAuditLogResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryAuditLogResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAuditLogResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut audit_logs__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AuditLogs => {
                            if audit_logs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditLogs"));
                            }
                            audit_logs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAuditLogResponse {
                    audit_logs: audit_logs__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryAuditLogResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.report_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportRequest", len)?;
        if !self.report_id.is_empty() {
            struct_ser.serialize_field("reportId", &self.report_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "report_id",
            "reportId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReportId,
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
                            "reportId" | "report_id" => Ok(GeneratedField::ReportId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut report_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReportId => {
                            if report_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportId"));
                            }
                            report_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryFraudReportRequest {
                    report_id: report_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.report.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportResponse", len)?;
        if let Some(v) = self.report.as_ref() {
            struct_ser.serialize_field("report", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "report",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Report,
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
                            "report" => Ok(GeneratedField::Report),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut report__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Report => {
                            if report__.is_some() {
                                return Err(serde::de::Error::duplicate_field("report"));
                            }
                            report__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFraudReportResponse {
                    report: report__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportsByReportedPartyRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reported_party.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportsByReportedPartyRequest", len)?;
        if !self.reported_party.is_empty() {
            struct_ser.serialize_field("reportedParty", &self.reported_party)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportsByReportedPartyRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reported_party",
            "reportedParty",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReportedParty,
            Pagination,
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
                            "reportedParty" | "reported_party" => Ok(GeneratedField::ReportedParty),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportsByReportedPartyRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportsByReportedPartyRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportsByReportedPartyRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reported_party__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReportedParty => {
                            if reported_party__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportedParty"));
                            }
                            reported_party__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFraudReportsByReportedPartyRequest {
                    reported_party: reported_party__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportsByReportedPartyRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportsByReportedPartyResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reports.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportsByReportedPartyResponse", len)?;
        if !self.reports.is_empty() {
            struct_ser.serialize_field("reports", &self.reports)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportsByReportedPartyResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reports",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reports,
            Pagination,
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
                            "reports" => Ok(GeneratedField::Reports),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportsByReportedPartyResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportsByReportedPartyResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportsByReportedPartyResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reports__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reports => {
                            if reports__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reports"));
                            }
                            reports__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFraudReportsByReportedPartyResponse {
                    reports: reports__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportsByReportedPartyResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportsByReporterRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reporter.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportsByReporterRequest", len)?;
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", &self.reporter)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportsByReporterRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reporter",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reporter,
            Pagination,
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
                            "reporter" => Ok(GeneratedField::Reporter),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportsByReporterRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportsByReporterRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportsByReporterRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reporter__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFraudReportsByReporterRequest {
                    reporter: reporter__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportsByReporterRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportsByReporterResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reports.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportsByReporterResponse", len)?;
        if !self.reports.is_empty() {
            struct_ser.serialize_field("reports", &self.reports)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportsByReporterResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reports",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reports,
            Pagination,
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
                            "reports" => Ok(GeneratedField::Reports),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportsByReporterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportsByReporterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportsByReporterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reports__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reports => {
                            if reports__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reports"));
                            }
                            reports__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFraudReportsByReporterResponse {
                    reports: reports__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportsByReporterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.pagination.is_some() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.category != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportsRequest", len)?;
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        if self.status != 0 {
            let v = FraudReportStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.category != 0 {
            let v = FraudCategory::try_from(self.category)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.category)))?;
            struct_ser.serialize_field("category", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "pagination",
            "status",
            "category",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Pagination,
            Status,
            Category,
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
                            "pagination" => Ok(GeneratedField::Pagination),
                            "status" => Ok(GeneratedField::Status),
                            "category" => Ok(GeneratedField::Category),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut pagination__ = None;
                let mut status__ = None;
                let mut category__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<FraudReportStatus>()? as i32);
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value::<FraudCategory>()? as i32);
                        }
                    }
                }
                Ok(QueryFraudReportsRequest {
                    pagination: pagination__,
                    status: status__.unwrap_or_default(),
                    category: category__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFraudReportsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reports.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryFraudReportsResponse", len)?;
        if !self.reports.is_empty() {
            struct_ser.serialize_field("reports", &self.reports)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFraudReportsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reports",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reports,
            Pagination,
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
                            "reports" => Ok(GeneratedField::Reports),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFraudReportsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryFraudReportsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFraudReportsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reports__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reports => {
                            if reports__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reports"));
                            }
                            reports__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFraudReportsResponse {
                    reports: reports__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryFraudReportsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModeratorQueueRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.pagination.is_some() {
            len += 1;
        }
        if self.category != 0 {
            len += 1;
        }
        if !self.assigned_to.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryModeratorQueueRequest", len)?;
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        if self.category != 0 {
            let v = FraudCategory::try_from(self.category)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.category)))?;
            struct_ser.serialize_field("category", &v)?;
        }
        if !self.assigned_to.is_empty() {
            struct_ser.serialize_field("assignedTo", &self.assigned_to)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModeratorQueueRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "pagination",
            "category",
            "assigned_to",
            "assignedTo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Pagination,
            Category,
            AssignedTo,
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
                            "pagination" => Ok(GeneratedField::Pagination),
                            "category" => Ok(GeneratedField::Category),
                            "assignedTo" | "assigned_to" => Ok(GeneratedField::AssignedTo),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryModeratorQueueRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryModeratorQueueRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModeratorQueueRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut pagination__ = None;
                let mut category__ = None;
                let mut assigned_to__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value::<FraudCategory>()? as i32);
                        }
                        GeneratedField::AssignedTo => {
                            if assigned_to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedTo"));
                            }
                            assigned_to__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryModeratorQueueRequest {
                    pagination: pagination__,
                    category: category__.unwrap_or_default(),
                    assigned_to: assigned_to__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryModeratorQueueRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModeratorQueueResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.queue_entries.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryModeratorQueueResponse", len)?;
        if !self.queue_entries.is_empty() {
            struct_ser.serialize_field("queueEntries", &self.queue_entries)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModeratorQueueResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "queue_entries",
            "queueEntries",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            QueueEntries,
            Pagination,
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
                            "queueEntries" | "queue_entries" => Ok(GeneratedField::QueueEntries),
                            "pagination" => Ok(GeneratedField::Pagination),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryModeratorQueueResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryModeratorQueueResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModeratorQueueResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut queue_entries__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::QueueEntries => {
                            if queue_entries__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queueEntries"));
                            }
                            queue_entries__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryModeratorQueueResponse {
                    queue_entries: queue_entries__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryModeratorQueueResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryParamsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryParamsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryParamsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
                            Err(serde::de::Error::unknown_field(value, FIELDS))
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryParamsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryParamsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryParamsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryParamsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryParamsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.params.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.fraud.v1.QueryParamsResponse", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryParamsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "params",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
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
                            "params" => Ok(GeneratedField::Params),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.fraud.v1.QueryParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryParamsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryParamsResponse {
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.fraud.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ResolutionType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "RESOLUTION_TYPE_UNSPECIFIED",
            Self::Warning => "RESOLUTION_TYPE_WARNING",
            Self::Suspension => "RESOLUTION_TYPE_SUSPENSION",
            Self::Termination => "RESOLUTION_TYPE_TERMINATION",
            Self::Refund => "RESOLUTION_TYPE_REFUND",
            Self::NoAction => "RESOLUTION_TYPE_NO_ACTION",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ResolutionType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "RESOLUTION_TYPE_UNSPECIFIED",
            "RESOLUTION_TYPE_WARNING",
            "RESOLUTION_TYPE_SUSPENSION",
            "RESOLUTION_TYPE_TERMINATION",
            "RESOLUTION_TYPE_REFUND",
            "RESOLUTION_TYPE_NO_ACTION",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ResolutionType;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                write!(formatter, "expected one of: {:?}", &FIELDS)
            }

            fn visit_i64<E>(self, v: i64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Signed(v), &self)
                    })
            }

            fn visit_u64<E>(self, v: u64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                i32::try_from(v)
                    .ok()
                    .and_then(|x| x.try_into().ok())
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "RESOLUTION_TYPE_UNSPECIFIED" => Ok(ResolutionType::Unspecified),
                    "RESOLUTION_TYPE_WARNING" => Ok(ResolutionType::Warning),
                    "RESOLUTION_TYPE_SUSPENSION" => Ok(ResolutionType::Suspension),
                    "RESOLUTION_TYPE_TERMINATION" => Ok(ResolutionType::Termination),
                    "RESOLUTION_TYPE_REFUND" => Ok(ResolutionType::Refund),
                    "RESOLUTION_TYPE_NO_ACTION" => Ok(ResolutionType::NoAction),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}

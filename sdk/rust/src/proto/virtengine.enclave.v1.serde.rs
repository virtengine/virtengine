// @generated
impl serde::Serialize for AttestedScoringResult {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.score != 0 {
            len += 1;
        }
        if !self.status.is_empty() {
            len += 1;
        }
        if !self.reason_codes.is_empty() {
            len += 1;
        }
        if !self.model_version_hash.is_empty() {
            len += 1;
        }
        if !self.input_hash.is_empty() {
            len += 1;
        }
        if !self.evidence_hashes.is_empty() {
            len += 1;
        }
        if !self.enclave_measurement_hash.is_empty() {
            len += 1;
        }
        if !self.enclave_signature.is_empty() {
            len += 1;
        }
        if !self.attestation_reference.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.timestamp.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.AttestedScoringResult", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.score != 0 {
            struct_ser.serialize_field("score", &self.score)?;
        }
        if !self.status.is_empty() {
            struct_ser.serialize_field("status", &self.status)?;
        }
        if !self.reason_codes.is_empty() {
            struct_ser.serialize_field("reasonCodes", &self.reason_codes)?;
        }
        if !self.model_version_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("modelVersionHash", pbjson::private::base64::encode(&self.model_version_hash).as_str())?;
        }
        if !self.input_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("inputHash", pbjson::private::base64::encode(&self.input_hash).as_str())?;
        }
        if !self.evidence_hashes.is_empty() {
            struct_ser.serialize_field("evidenceHashes", &self.evidence_hashes.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.enclave_measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("enclaveMeasurementHash", pbjson::private::base64::encode(&self.enclave_measurement_hash).as_str())?;
        }
        if !self.enclave_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("enclaveSignature", pbjson::private::base64::encode(&self.enclave_signature).as_str())?;
        }
        if !self.attestation_reference.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("attestationReference", pbjson::private::base64::encode(&self.attestation_reference).as_str())?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if let Some(v) = self.timestamp.as_ref() {
            struct_ser.serialize_field("timestamp", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AttestedScoringResult {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "account_address",
            "accountAddress",
            "score",
            "status",
            "reason_codes",
            "reasonCodes",
            "model_version_hash",
            "modelVersionHash",
            "input_hash",
            "inputHash",
            "evidence_hashes",
            "evidenceHashes",
            "enclave_measurement_hash",
            "enclaveMeasurementHash",
            "enclave_signature",
            "enclaveSignature",
            "attestation_reference",
            "attestationReference",
            "validator_address",
            "validatorAddress",
            "block_height",
            "blockHeight",
            "timestamp",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            AccountAddress,
            Score,
            Status,
            ReasonCodes,
            ModelVersionHash,
            InputHash,
            EvidenceHashes,
            EnclaveMeasurementHash,
            EnclaveSignature,
            AttestationReference,
            ValidatorAddress,
            BlockHeight,
            Timestamp,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "score" => Ok(GeneratedField::Score),
                            "status" => Ok(GeneratedField::Status),
                            "reasonCodes" | "reason_codes" => Ok(GeneratedField::ReasonCodes),
                            "modelVersionHash" | "model_version_hash" => Ok(GeneratedField::ModelVersionHash),
                            "inputHash" | "input_hash" => Ok(GeneratedField::InputHash),
                            "evidenceHashes" | "evidence_hashes" => Ok(GeneratedField::EvidenceHashes),
                            "enclaveMeasurementHash" | "enclave_measurement_hash" => Ok(GeneratedField::EnclaveMeasurementHash),
                            "enclaveSignature" | "enclave_signature" => Ok(GeneratedField::EnclaveSignature),
                            "attestationReference" | "attestation_reference" => Ok(GeneratedField::AttestationReference),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "timestamp" => Ok(GeneratedField::Timestamp),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AttestedScoringResult;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.AttestedScoringResult")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AttestedScoringResult, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut account_address__ = None;
                let mut score__ = None;
                let mut status__ = None;
                let mut reason_codes__ = None;
                let mut model_version_hash__ = None;
                let mut input_hash__ = None;
                let mut evidence_hashes__ = None;
                let mut enclave_measurement_hash__ = None;
                let mut enclave_signature__ = None;
                let mut attestation_reference__ = None;
                let mut validator_address__ = None;
                let mut block_height__ = None;
                let mut timestamp__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReasonCodes => {
                            if reason_codes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reasonCodes"));
                            }
                            reason_codes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelVersionHash => {
                            if model_version_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersionHash"));
                            }
                            model_version_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::InputHash => {
                            if input_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("inputHash"));
                            }
                            input_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EvidenceHashes => {
                            if evidence_hashes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidenceHashes"));
                            }
                            evidence_hashes__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::EnclaveMeasurementHash => {
                            if enclave_measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enclaveMeasurementHash"));
                            }
                            enclave_measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EnclaveSignature => {
                            if enclave_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enclaveSignature"));
                            }
                            enclave_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationReference => {
                            if attestation_reference__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationReference"));
                            }
                            attestation_reference__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = map_.next_value()?;
                        }
                    }
                }
                Ok(AttestedScoringResult {
                    scope_id: scope_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    score: score__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    reason_codes: reason_codes__.unwrap_or_default(),
                    model_version_hash: model_version_hash__.unwrap_or_default(),
                    input_hash: input_hash__.unwrap_or_default(),
                    evidence_hashes: evidence_hashes__.unwrap_or_default(),
                    enclave_measurement_hash: enclave_measurement_hash__.unwrap_or_default(),
                    enclave_signature: enclave_signature__.unwrap_or_default(),
                    attestation_reference: attestation_reference__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    timestamp: timestamp__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.AttestedScoringResult", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EnclaveIdentity {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.tee_type != 0 {
            len += 1;
        }
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if !self.signer_hash.is_empty() {
            len += 1;
        }
        if !self.encryption_pub_key.is_empty() {
            len += 1;
        }
        if !self.signing_pub_key.is_empty() {
            len += 1;
        }
        if !self.attestation_quote.is_empty() {
            len += 1;
        }
        if !self.attestation_chain.is_empty() {
            len += 1;
        }
        if self.isv_prod_id != 0 {
            len += 1;
        }
        if self.isv_svn != 0 {
            len += 1;
        }
        if self.quote_version != 0 {
            len += 1;
        }
        if self.debug_mode {
            len += 1;
        }
        if self.epoch != 0 {
            len += 1;
        }
        if self.expiry_height != 0 {
            len += 1;
        }
        if self.registered_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EnclaveIdentity", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.tee_type != 0 {
            let v = TeeType::try_from(self.tee_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tee_type)))?;
            struct_ser.serialize_field("teeType", &v)?;
        }
        if !self.measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("measurementHash", pbjson::private::base64::encode(&self.measurement_hash).as_str())?;
        }
        if !self.signer_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signerHash", pbjson::private::base64::encode(&self.signer_hash).as_str())?;
        }
        if !self.encryption_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("encryptionPubKey", pbjson::private::base64::encode(&self.encryption_pub_key).as_str())?;
        }
        if !self.signing_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signingPubKey", pbjson::private::base64::encode(&self.signing_pub_key).as_str())?;
        }
        if !self.attestation_quote.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("attestationQuote", pbjson::private::base64::encode(&self.attestation_quote).as_str())?;
        }
        if !self.attestation_chain.is_empty() {
            struct_ser.serialize_field("attestationChain", &self.attestation_chain.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if self.isv_prod_id != 0 {
            struct_ser.serialize_field("isvProdId", &self.isv_prod_id)?;
        }
        if self.isv_svn != 0 {
            struct_ser.serialize_field("isvSvn", &self.isv_svn)?;
        }
        if self.quote_version != 0 {
            struct_ser.serialize_field("quoteVersion", &self.quote_version)?;
        }
        if self.debug_mode {
            struct_ser.serialize_field("debugMode", &self.debug_mode)?;
        }
        if self.epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epoch", ToString::to_string(&self.epoch).as_str())?;
        }
        if self.expiry_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryHeight", ToString::to_string(&self.expiry_height).as_str())?;
        }
        if let Some(v) = self.registered_at.as_ref() {
            struct_ser.serialize_field("registeredAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if self.status != 0 {
            let v = EnclaveIdentityStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EnclaveIdentity {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "tee_type",
            "teeType",
            "measurement_hash",
            "measurementHash",
            "signer_hash",
            "signerHash",
            "encryption_pub_key",
            "encryptionPubKey",
            "signing_pub_key",
            "signingPubKey",
            "attestation_quote",
            "attestationQuote",
            "attestation_chain",
            "attestationChain",
            "isv_prod_id",
            "isvProdId",
            "isv_svn",
            "isvSvn",
            "quote_version",
            "quoteVersion",
            "debug_mode",
            "debugMode",
            "epoch",
            "expiry_height",
            "expiryHeight",
            "registered_at",
            "registeredAt",
            "updated_at",
            "updatedAt",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            TeeType,
            MeasurementHash,
            SignerHash,
            EncryptionPubKey,
            SigningPubKey,
            AttestationQuote,
            AttestationChain,
            IsvProdId,
            IsvSvn,
            QuoteVersion,
            DebugMode,
            Epoch,
            ExpiryHeight,
            RegisteredAt,
            UpdatedAt,
            Status,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "signerHash" | "signer_hash" => Ok(GeneratedField::SignerHash),
                            "encryptionPubKey" | "encryption_pub_key" => Ok(GeneratedField::EncryptionPubKey),
                            "signingPubKey" | "signing_pub_key" => Ok(GeneratedField::SigningPubKey),
                            "attestationQuote" | "attestation_quote" => Ok(GeneratedField::AttestationQuote),
                            "attestationChain" | "attestation_chain" => Ok(GeneratedField::AttestationChain),
                            "isvProdId" | "isv_prod_id" => Ok(GeneratedField::IsvProdId),
                            "isvSvn" | "isv_svn" => Ok(GeneratedField::IsvSvn),
                            "quoteVersion" | "quote_version" => Ok(GeneratedField::QuoteVersion),
                            "debugMode" | "debug_mode" => Ok(GeneratedField::DebugMode),
                            "epoch" => Ok(GeneratedField::Epoch),
                            "expiryHeight" | "expiry_height" => Ok(GeneratedField::ExpiryHeight),
                            "registeredAt" | "registered_at" => Ok(GeneratedField::RegisteredAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EnclaveIdentity;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EnclaveIdentity")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EnclaveIdentity, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut tee_type__ = None;
                let mut measurement_hash__ = None;
                let mut signer_hash__ = None;
                let mut encryption_pub_key__ = None;
                let mut signing_pub_key__ = None;
                let mut attestation_quote__ = None;
                let mut attestation_chain__ = None;
                let mut isv_prod_id__ = None;
                let mut isv_svn__ = None;
                let mut quote_version__ = None;
                let mut debug_mode__ = None;
                let mut epoch__ = None;
                let mut expiry_height__ = None;
                let mut registered_at__ = None;
                let mut updated_at__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value::<TeeType>()? as i32);
                        }
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SignerHash => {
                            if signer_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signerHash"));
                            }
                            signer_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EncryptionPubKey => {
                            if encryption_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptionPubKey"));
                            }
                            encryption_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SigningPubKey => {
                            if signing_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signingPubKey"));
                            }
                            signing_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationQuote => {
                            if attestation_quote__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationQuote"));
                            }
                            attestation_quote__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationChain => {
                            if attestation_chain__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationChain"));
                            }
                            attestation_chain__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::IsvProdId => {
                            if isv_prod_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isvProdId"));
                            }
                            isv_prod_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsvSvn => {
                            if isv_svn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isvSvn"));
                            }
                            isv_svn__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::QuoteVersion => {
                            if quote_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quoteVersion"));
                            }
                            quote_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DebugMode => {
                            if debug_mode__.is_some() {
                                return Err(serde::de::Error::duplicate_field("debugMode"));
                            }
                            debug_mode__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Epoch => {
                            if epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epoch"));
                            }
                            epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiryHeight => {
                            if expiry_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryHeight"));
                            }
                            expiry_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RegisteredAt => {
                            if registered_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registeredAt"));
                            }
                            registered_at__ = map_.next_value()?;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<EnclaveIdentityStatus>()? as i32);
                        }
                    }
                }
                Ok(EnclaveIdentity {
                    validator_address: validator_address__.unwrap_or_default(),
                    tee_type: tee_type__.unwrap_or_default(),
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    signer_hash: signer_hash__.unwrap_or_default(),
                    encryption_pub_key: encryption_pub_key__.unwrap_or_default(),
                    signing_pub_key: signing_pub_key__.unwrap_or_default(),
                    attestation_quote: attestation_quote__.unwrap_or_default(),
                    attestation_chain: attestation_chain__.unwrap_or_default(),
                    isv_prod_id: isv_prod_id__.unwrap_or_default(),
                    isv_svn: isv_svn__.unwrap_or_default(),
                    quote_version: quote_version__.unwrap_or_default(),
                    debug_mode: debug_mode__.unwrap_or_default(),
                    epoch: epoch__.unwrap_or_default(),
                    expiry_height: expiry_height__.unwrap_or_default(),
                    registered_at: registered_at__,
                    updated_at: updated_at__,
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EnclaveIdentity", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EnclaveIdentityStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "ENCLAVE_IDENTITY_STATUS_UNSPECIFIED",
            Self::Active => "ENCLAVE_IDENTITY_STATUS_ACTIVE",
            Self::Pending => "ENCLAVE_IDENTITY_STATUS_PENDING",
            Self::Expired => "ENCLAVE_IDENTITY_STATUS_EXPIRED",
            Self::Revoked => "ENCLAVE_IDENTITY_STATUS_REVOKED",
            Self::Rotating => "ENCLAVE_IDENTITY_STATUS_ROTATING",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for EnclaveIdentityStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ENCLAVE_IDENTITY_STATUS_UNSPECIFIED",
            "ENCLAVE_IDENTITY_STATUS_ACTIVE",
            "ENCLAVE_IDENTITY_STATUS_PENDING",
            "ENCLAVE_IDENTITY_STATUS_EXPIRED",
            "ENCLAVE_IDENTITY_STATUS_REVOKED",
            "ENCLAVE_IDENTITY_STATUS_ROTATING",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EnclaveIdentityStatus;

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
                    "ENCLAVE_IDENTITY_STATUS_UNSPECIFIED" => Ok(EnclaveIdentityStatus::Unspecified),
                    "ENCLAVE_IDENTITY_STATUS_ACTIVE" => Ok(EnclaveIdentityStatus::Active),
                    "ENCLAVE_IDENTITY_STATUS_PENDING" => Ok(EnclaveIdentityStatus::Pending),
                    "ENCLAVE_IDENTITY_STATUS_EXPIRED" => Ok(EnclaveIdentityStatus::Expired),
                    "ENCLAVE_IDENTITY_STATUS_REVOKED" => Ok(EnclaveIdentityStatus::Revoked),
                    "ENCLAVE_IDENTITY_STATUS_ROTATING" => Ok(EnclaveIdentityStatus::Rotating),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for EventConsensusVerificationFailed {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.proposed_score != 0 {
            len += 1;
        }
        if self.computed_score != 0 {
            len += 1;
        }
        if self.score_difference != 0 {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventConsensusVerificationFailed", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.proposed_score != 0 {
            struct_ser.serialize_field("proposedScore", &self.proposed_score)?;
        }
        if self.computed_score != 0 {
            struct_ser.serialize_field("computedScore", &self.computed_score)?;
        }
        if self.score_difference != 0 {
            struct_ser.serialize_field("scoreDifference", &self.score_difference)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventConsensusVerificationFailed {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "proposed_score",
            "proposedScore",
            "computed_score",
            "computedScore",
            "score_difference",
            "scoreDifference",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            ProposedScore,
            ComputedScore,
            ScoreDifference,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "proposedScore" | "proposed_score" => Ok(GeneratedField::ProposedScore),
                            "computedScore" | "computed_score" => Ok(GeneratedField::ComputedScore),
                            "scoreDifference" | "score_difference" => Ok(GeneratedField::ScoreDifference),
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
            type Value = EventConsensusVerificationFailed;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventConsensusVerificationFailed")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventConsensusVerificationFailed, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut proposed_score__ = None;
                let mut computed_score__ = None;
                let mut score_difference__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProposedScore => {
                            if proposed_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposedScore"));
                            }
                            proposed_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ComputedScore => {
                            if computed_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("computedScore"));
                            }
                            computed_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScoreDifference => {
                            if score_difference__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreDifference"));
                            }
                            score_difference__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventConsensusVerificationFailed {
                    scope_id: scope_id__.unwrap_or_default(),
                    proposed_score: proposed_score__.unwrap_or_default(),
                    computed_score: computed_score__.unwrap_or_default(),
                    score_difference: score_difference__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventConsensusVerificationFailed", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEnclaveIdentityExpired {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator.is_empty() {
            len += 1;
        }
        if self.expiry_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventEnclaveIdentityExpired", len)?;
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if self.expiry_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryHeight", ToString::to_string(&self.expiry_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEnclaveIdentityExpired {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator",
            "expiry_height",
            "expiryHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Validator,
            ExpiryHeight,
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
                            "validator" => Ok(GeneratedField::Validator),
                            "expiryHeight" | "expiry_height" => Ok(GeneratedField::ExpiryHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventEnclaveIdentityExpired;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventEnclaveIdentityExpired")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEnclaveIdentityExpired, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator__ = None;
                let mut expiry_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiryHeight => {
                            if expiry_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryHeight"));
                            }
                            expiry_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventEnclaveIdentityExpired {
                    validator: validator__.unwrap_or_default(),
                    expiry_height: expiry_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventEnclaveIdentityExpired", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEnclaveIdentityRegistered {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator.is_empty() {
            len += 1;
        }
        if !self.tee_type.is_empty() {
            len += 1;
        }
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if !self.encryption_key_id.is_empty() {
            len += 1;
        }
        if !self.signing_key_id.is_empty() {
            len += 1;
        }
        if self.epoch != 0 {
            len += 1;
        }
        if self.expiry_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventEnclaveIdentityRegistered", len)?;
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if !self.tee_type.is_empty() {
            struct_ser.serialize_field("teeType", &self.tee_type)?;
        }
        if !self.measurement_hash.is_empty() {
            struct_ser.serialize_field("measurementHash", &self.measurement_hash)?;
        }
        if !self.encryption_key_id.is_empty() {
            struct_ser.serialize_field("encryptionKeyId", &self.encryption_key_id)?;
        }
        if !self.signing_key_id.is_empty() {
            struct_ser.serialize_field("signingKeyId", &self.signing_key_id)?;
        }
        if self.epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epoch", ToString::to_string(&self.epoch).as_str())?;
        }
        if self.expiry_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryHeight", ToString::to_string(&self.expiry_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEnclaveIdentityRegistered {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator",
            "tee_type",
            "teeType",
            "measurement_hash",
            "measurementHash",
            "encryption_key_id",
            "encryptionKeyId",
            "signing_key_id",
            "signingKeyId",
            "epoch",
            "expiry_height",
            "expiryHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Validator,
            TeeType,
            MeasurementHash,
            EncryptionKeyId,
            SigningKeyId,
            Epoch,
            ExpiryHeight,
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
                            "validator" => Ok(GeneratedField::Validator),
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "encryptionKeyId" | "encryption_key_id" => Ok(GeneratedField::EncryptionKeyId),
                            "signingKeyId" | "signing_key_id" => Ok(GeneratedField::SigningKeyId),
                            "epoch" => Ok(GeneratedField::Epoch),
                            "expiryHeight" | "expiry_height" => Ok(GeneratedField::ExpiryHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventEnclaveIdentityRegistered;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventEnclaveIdentityRegistered")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEnclaveIdentityRegistered, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator__ = None;
                let mut tee_type__ = None;
                let mut measurement_hash__ = None;
                let mut encryption_key_id__ = None;
                let mut signing_key_id__ = None;
                let mut epoch__ = None;
                let mut expiry_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptionKeyId => {
                            if encryption_key_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptionKeyId"));
                            }
                            encryption_key_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SigningKeyId => {
                            if signing_key_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signingKeyId"));
                            }
                            signing_key_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Epoch => {
                            if epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epoch"));
                            }
                            epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiryHeight => {
                            if expiry_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryHeight"));
                            }
                            expiry_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventEnclaveIdentityRegistered {
                    validator: validator__.unwrap_or_default(),
                    tee_type: tee_type__.unwrap_or_default(),
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    encryption_key_id: encryption_key_id__.unwrap_or_default(),
                    signing_key_id: signing_key_id__.unwrap_or_default(),
                    epoch: epoch__.unwrap_or_default(),
                    expiry_height: expiry_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventEnclaveIdentityRegistered", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEnclaveIdentityRevoked {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventEnclaveIdentityRevoked", len)?;
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEnclaveIdentityRevoked {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Validator,
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
                            "validator" => Ok(GeneratedField::Validator),
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
            type Value = EventEnclaveIdentityRevoked;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventEnclaveIdentityRevoked")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEnclaveIdentityRevoked, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventEnclaveIdentityRevoked {
                    validator: validator__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventEnclaveIdentityRevoked", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEnclaveIdentityUpdated {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator.is_empty() {
            len += 1;
        }
        if !self.status.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventEnclaveIdentityUpdated", len)?;
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if !self.status.is_empty() {
            struct_ser.serialize_field("status", &self.status)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEnclaveIdentityUpdated {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Validator,
            Status,
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
                            "validator" => Ok(GeneratedField::Validator),
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventEnclaveIdentityUpdated;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventEnclaveIdentityUpdated")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEnclaveIdentityUpdated, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventEnclaveIdentityUpdated {
                    validator: validator__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventEnclaveIdentityUpdated", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEnclaveKeyRotated {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator.is_empty() {
            len += 1;
        }
        if !self.old_key_fingerprint.is_empty() {
            len += 1;
        }
        if !self.new_key_fingerprint.is_empty() {
            len += 1;
        }
        if self.overlap_start_height != 0 {
            len += 1;
        }
        if self.overlap_end_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventEnclaveKeyRotated", len)?;
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if !self.old_key_fingerprint.is_empty() {
            struct_ser.serialize_field("oldKeyFingerprint", &self.old_key_fingerprint)?;
        }
        if !self.new_key_fingerprint.is_empty() {
            struct_ser.serialize_field("newKeyFingerprint", &self.new_key_fingerprint)?;
        }
        if self.overlap_start_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapStartHeight", ToString::to_string(&self.overlap_start_height).as_str())?;
        }
        if self.overlap_end_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapEndHeight", ToString::to_string(&self.overlap_end_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEnclaveKeyRotated {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator",
            "old_key_fingerprint",
            "oldKeyFingerprint",
            "new_key_fingerprint",
            "newKeyFingerprint",
            "overlap_start_height",
            "overlapStartHeight",
            "overlap_end_height",
            "overlapEndHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Validator,
            OldKeyFingerprint,
            NewKeyFingerprint,
            OverlapStartHeight,
            OverlapEndHeight,
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
                            "validator" => Ok(GeneratedField::Validator),
                            "oldKeyFingerprint" | "old_key_fingerprint" => Ok(GeneratedField::OldKeyFingerprint),
                            "newKeyFingerprint" | "new_key_fingerprint" => Ok(GeneratedField::NewKeyFingerprint),
                            "overlapStartHeight" | "overlap_start_height" => Ok(GeneratedField::OverlapStartHeight),
                            "overlapEndHeight" | "overlap_end_height" => Ok(GeneratedField::OverlapEndHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventEnclaveKeyRotated;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventEnclaveKeyRotated")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEnclaveKeyRotated, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator__ = None;
                let mut old_key_fingerprint__ = None;
                let mut new_key_fingerprint__ = None;
                let mut overlap_start_height__ = None;
                let mut overlap_end_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OldKeyFingerprint => {
                            if old_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldKeyFingerprint"));
                            }
                            old_key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewKeyFingerprint => {
                            if new_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newKeyFingerprint"));
                            }
                            new_key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OverlapStartHeight => {
                            if overlap_start_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapStartHeight"));
                            }
                            overlap_start_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OverlapEndHeight => {
                            if overlap_end_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapEndHeight"));
                            }
                            overlap_end_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventEnclaveKeyRotated {
                    validator: validator__.unwrap_or_default(),
                    old_key_fingerprint: old_key_fingerprint__.unwrap_or_default(),
                    new_key_fingerprint: new_key_fingerprint__.unwrap_or_default(),
                    overlap_start_height: overlap_start_height__.unwrap_or_default(),
                    overlap_end_height: overlap_end_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventEnclaveKeyRotated", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventKeyRotationCompleted {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator.is_empty() {
            len += 1;
        }
        if !self.new_key_fingerprint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventKeyRotationCompleted", len)?;
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if !self.new_key_fingerprint.is_empty() {
            struct_ser.serialize_field("newKeyFingerprint", &self.new_key_fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventKeyRotationCompleted {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator",
            "new_key_fingerprint",
            "newKeyFingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Validator,
            NewKeyFingerprint,
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
                            "validator" => Ok(GeneratedField::Validator),
                            "newKeyFingerprint" | "new_key_fingerprint" => Ok(GeneratedField::NewKeyFingerprint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventKeyRotationCompleted;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventKeyRotationCompleted")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventKeyRotationCompleted, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator__ = None;
                let mut new_key_fingerprint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewKeyFingerprint => {
                            if new_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newKeyFingerprint"));
                            }
                            new_key_fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventKeyRotationCompleted {
                    validator: validator__.unwrap_or_default(),
                    new_key_fingerprint: new_key_fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventKeyRotationCompleted", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventMeasurementAdded {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if !self.tee_type.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.min_isv_svn != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventMeasurementAdded", len)?;
        if !self.measurement_hash.is_empty() {
            struct_ser.serialize_field("measurementHash", &self.measurement_hash)?;
        }
        if !self.tee_type.is_empty() {
            struct_ser.serialize_field("teeType", &self.tee_type)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.min_isv_svn != 0 {
            struct_ser.serialize_field("minIsvSvn", &self.min_isv_svn)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventMeasurementAdded {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurement_hash",
            "measurementHash",
            "tee_type",
            "teeType",
            "description",
            "min_isv_svn",
            "minIsvSvn",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MeasurementHash,
            TeeType,
            Description,
            MinIsvSvn,
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
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "description" => Ok(GeneratedField::Description),
                            "minIsvSvn" | "min_isv_svn" => Ok(GeneratedField::MinIsvSvn),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventMeasurementAdded;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventMeasurementAdded")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventMeasurementAdded, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurement_hash__ = None;
                let mut tee_type__ = None;
                let mut description__ = None;
                let mut min_isv_svn__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinIsvSvn => {
                            if min_isv_svn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minIsvSvn"));
                            }
                            min_isv_svn__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventMeasurementAdded {
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    tee_type: tee_type__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    min_isv_svn: min_isv_svn__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventMeasurementAdded", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventMeasurementRevoked {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventMeasurementRevoked", len)?;
        if !self.measurement_hash.is_empty() {
            struct_ser.serialize_field("measurementHash", &self.measurement_hash)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventMeasurementRevoked {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurement_hash",
            "measurementHash",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MeasurementHash,
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
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
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
            type Value = EventMeasurementRevoked;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventMeasurementRevoked")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventMeasurementRevoked, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurement_hash__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventMeasurementRevoked {
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventMeasurementRevoked", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventVeidScoreComputedAttested {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.score != 0 {
            len += 1;
        }
        if !self.status.is_empty() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventVEIDScoreComputedAttested", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.score != 0 {
            struct_ser.serialize_field("score", &self.score)?;
        }
        if !self.status.is_empty() {
            struct_ser.serialize_field("status", &self.status)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventVeidScoreComputedAttested {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "account_address",
            "accountAddress",
            "score",
            "status",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            AccountAddress,
            Score,
            Status,
            BlockHeight,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "score" => Ok(GeneratedField::Score),
                            "status" => Ok(GeneratedField::Status),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventVeidScoreComputedAttested;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventVEIDScoreComputedAttested")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventVeidScoreComputedAttested, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut account_address__ = None;
                let mut score__ = None;
                let mut status__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventVeidScoreComputedAttested {
                    scope_id: scope_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    score: score__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventVEIDScoreComputedAttested", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventVeidScoreRejectedAttestation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.proposed_score != 0 {
            len += 1;
        }
        if self.computed_score != 0 {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.EventVEIDScoreRejectedAttestation", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.proposed_score != 0 {
            struct_ser.serialize_field("proposedScore", &self.proposed_score)?;
        }
        if self.computed_score != 0 {
            struct_ser.serialize_field("computedScore", &self.computed_score)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventVeidScoreRejectedAttestation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "account_address",
            "accountAddress",
            "proposed_score",
            "proposedScore",
            "computed_score",
            "computedScore",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            AccountAddress,
            ProposedScore,
            ComputedScore,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "proposedScore" | "proposed_score" => Ok(GeneratedField::ProposedScore),
                            "computedScore" | "computed_score" => Ok(GeneratedField::ComputedScore),
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
            type Value = EventVeidScoreRejectedAttestation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.EventVEIDScoreRejectedAttestation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventVeidScoreRejectedAttestation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut account_address__ = None;
                let mut proposed_score__ = None;
                let mut computed_score__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProposedScore => {
                            if proposed_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposedScore"));
                            }
                            proposed_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ComputedScore => {
                            if computed_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("computedScore"));
                            }
                            computed_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventVeidScoreRejectedAttestation {
                    scope_id: scope_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    proposed_score: proposed_score__.unwrap_or_default(),
                    computed_score: computed_score__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.EventVEIDScoreRejectedAttestation", FIELDS, GeneratedVisitor)
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
        if !self.enclave_identities.is_empty() {
            len += 1;
        }
        if !self.measurement_allowlist.is_empty() {
            len += 1;
        }
        if !self.key_rotations.is_empty() {
            len += 1;
        }
        if self.params.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.GenesisState", len)?;
        if !self.enclave_identities.is_empty() {
            struct_ser.serialize_field("enclaveIdentities", &self.enclave_identities)?;
        }
        if !self.measurement_allowlist.is_empty() {
            struct_ser.serialize_field("measurementAllowlist", &self.measurement_allowlist)?;
        }
        if !self.key_rotations.is_empty() {
            struct_ser.serialize_field("keyRotations", &self.key_rotations)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
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
            "enclave_identities",
            "enclaveIdentities",
            "measurement_allowlist",
            "measurementAllowlist",
            "key_rotations",
            "keyRotations",
            "params",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EnclaveIdentities,
            MeasurementAllowlist,
            KeyRotations,
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
                            "enclaveIdentities" | "enclave_identities" => Ok(GeneratedField::EnclaveIdentities),
                            "measurementAllowlist" | "measurement_allowlist" => Ok(GeneratedField::MeasurementAllowlist),
                            "keyRotations" | "key_rotations" => Ok(GeneratedField::KeyRotations),
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
            type Value = GenesisState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut enclave_identities__ = None;
                let mut measurement_allowlist__ = None;
                let mut key_rotations__ = None;
                let mut params__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EnclaveIdentities => {
                            if enclave_identities__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enclaveIdentities"));
                            }
                            enclave_identities__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MeasurementAllowlist => {
                            if measurement_allowlist__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementAllowlist"));
                            }
                            measurement_allowlist__ = Some(map_.next_value()?);
                        }
                        GeneratedField::KeyRotations => {
                            if key_rotations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyRotations"));
                            }
                            key_rotations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisState {
                    enclave_identities: enclave_identities__.unwrap_or_default(),
                    measurement_allowlist: measurement_allowlist__.unwrap_or_default(),
                    key_rotations: key_rotations__.unwrap_or_default(),
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for KeyRotationRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.epoch != 0 {
            len += 1;
        }
        if !self.old_key_fingerprint.is_empty() {
            len += 1;
        }
        if !self.new_key_fingerprint.is_empty() {
            len += 1;
        }
        if self.overlap_start_height != 0 {
            len += 1;
        }
        if self.overlap_end_height != 0 {
            len += 1;
        }
        if self.initiated_at.is_some() {
            len += 1;
        }
        if self.completed_at.is_some() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.KeyRotationRecord", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epoch", ToString::to_string(&self.epoch).as_str())?;
        }
        if !self.old_key_fingerprint.is_empty() {
            struct_ser.serialize_field("oldKeyFingerprint", &self.old_key_fingerprint)?;
        }
        if !self.new_key_fingerprint.is_empty() {
            struct_ser.serialize_field("newKeyFingerprint", &self.new_key_fingerprint)?;
        }
        if self.overlap_start_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapStartHeight", ToString::to_string(&self.overlap_start_height).as_str())?;
        }
        if self.overlap_end_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapEndHeight", ToString::to_string(&self.overlap_end_height).as_str())?;
        }
        if let Some(v) = self.initiated_at.as_ref() {
            struct_ser.serialize_field("initiatedAt", v)?;
        }
        if let Some(v) = self.completed_at.as_ref() {
            struct_ser.serialize_field("completedAt", v)?;
        }
        if self.status != 0 {
            let v = KeyRotationStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for KeyRotationRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "epoch",
            "old_key_fingerprint",
            "oldKeyFingerprint",
            "new_key_fingerprint",
            "newKeyFingerprint",
            "overlap_start_height",
            "overlapStartHeight",
            "overlap_end_height",
            "overlapEndHeight",
            "initiated_at",
            "initiatedAt",
            "completed_at",
            "completedAt",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            Epoch,
            OldKeyFingerprint,
            NewKeyFingerprint,
            OverlapStartHeight,
            OverlapEndHeight,
            InitiatedAt,
            CompletedAt,
            Status,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "epoch" => Ok(GeneratedField::Epoch),
                            "oldKeyFingerprint" | "old_key_fingerprint" => Ok(GeneratedField::OldKeyFingerprint),
                            "newKeyFingerprint" | "new_key_fingerprint" => Ok(GeneratedField::NewKeyFingerprint),
                            "overlapStartHeight" | "overlap_start_height" => Ok(GeneratedField::OverlapStartHeight),
                            "overlapEndHeight" | "overlap_end_height" => Ok(GeneratedField::OverlapEndHeight),
                            "initiatedAt" | "initiated_at" => Ok(GeneratedField::InitiatedAt),
                            "completedAt" | "completed_at" => Ok(GeneratedField::CompletedAt),
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = KeyRotationRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.KeyRotationRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<KeyRotationRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut epoch__ = None;
                let mut old_key_fingerprint__ = None;
                let mut new_key_fingerprint__ = None;
                let mut overlap_start_height__ = None;
                let mut overlap_end_height__ = None;
                let mut initiated_at__ = None;
                let mut completed_at__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Epoch => {
                            if epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epoch"));
                            }
                            epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OldKeyFingerprint => {
                            if old_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldKeyFingerprint"));
                            }
                            old_key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewKeyFingerprint => {
                            if new_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newKeyFingerprint"));
                            }
                            new_key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OverlapStartHeight => {
                            if overlap_start_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapStartHeight"));
                            }
                            overlap_start_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OverlapEndHeight => {
                            if overlap_end_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapEndHeight"));
                            }
                            overlap_end_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::InitiatedAt => {
                            if initiated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("initiatedAt"));
                            }
                            initiated_at__ = map_.next_value()?;
                        }
                        GeneratedField::CompletedAt => {
                            if completed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completedAt"));
                            }
                            completed_at__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<KeyRotationStatus>()? as i32);
                        }
                    }
                }
                Ok(KeyRotationRecord {
                    validator_address: validator_address__.unwrap_or_default(),
                    epoch: epoch__.unwrap_or_default(),
                    old_key_fingerprint: old_key_fingerprint__.unwrap_or_default(),
                    new_key_fingerprint: new_key_fingerprint__.unwrap_or_default(),
                    overlap_start_height: overlap_start_height__.unwrap_or_default(),
                    overlap_end_height: overlap_end_height__.unwrap_or_default(),
                    initiated_at: initiated_at__,
                    completed_at: completed_at__,
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.KeyRotationRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for KeyRotationStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "KEY_ROTATION_STATUS_UNSPECIFIED",
            Self::Pending => "KEY_ROTATION_STATUS_PENDING",
            Self::Active => "KEY_ROTATION_STATUS_ACTIVE",
            Self::Completed => "KEY_ROTATION_STATUS_COMPLETED",
            Self::Cancelled => "KEY_ROTATION_STATUS_CANCELLED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for KeyRotationStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "KEY_ROTATION_STATUS_UNSPECIFIED",
            "KEY_ROTATION_STATUS_PENDING",
            "KEY_ROTATION_STATUS_ACTIVE",
            "KEY_ROTATION_STATUS_COMPLETED",
            "KEY_ROTATION_STATUS_CANCELLED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = KeyRotationStatus;

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
                    "KEY_ROTATION_STATUS_UNSPECIFIED" => Ok(KeyRotationStatus::Unspecified),
                    "KEY_ROTATION_STATUS_PENDING" => Ok(KeyRotationStatus::Pending),
                    "KEY_ROTATION_STATUS_ACTIVE" => Ok(KeyRotationStatus::Active),
                    "KEY_ROTATION_STATUS_COMPLETED" => Ok(KeyRotationStatus::Completed),
                    "KEY_ROTATION_STATUS_CANCELLED" => Ok(KeyRotationStatus::Cancelled),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for MeasurementRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if self.tee_type != 0 {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.min_isv_svn != 0 {
            len += 1;
        }
        if self.added_at.is_some() {
            len += 1;
        }
        if self.added_by_proposal != 0 {
            len += 1;
        }
        if self.expiry_height != 0 {
            len += 1;
        }
        if self.revoked {
            len += 1;
        }
        if self.revoked_at.is_some() {
            len += 1;
        }
        if self.revoked_by_proposal != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MeasurementRecord", len)?;
        if !self.measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("measurementHash", pbjson::private::base64::encode(&self.measurement_hash).as_str())?;
        }
        if self.tee_type != 0 {
            let v = TeeType::try_from(self.tee_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tee_type)))?;
            struct_ser.serialize_field("teeType", &v)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.min_isv_svn != 0 {
            struct_ser.serialize_field("minIsvSvn", &self.min_isv_svn)?;
        }
        if let Some(v) = self.added_at.as_ref() {
            struct_ser.serialize_field("addedAt", v)?;
        }
        if self.added_by_proposal != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("addedByProposal", ToString::to_string(&self.added_by_proposal).as_str())?;
        }
        if self.expiry_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryHeight", ToString::to_string(&self.expiry_height).as_str())?;
        }
        if self.revoked {
            struct_ser.serialize_field("revoked", &self.revoked)?;
        }
        if let Some(v) = self.revoked_at.as_ref() {
            struct_ser.serialize_field("revokedAt", v)?;
        }
        if self.revoked_by_proposal != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedByProposal", ToString::to_string(&self.revoked_by_proposal).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MeasurementRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurement_hash",
            "measurementHash",
            "tee_type",
            "teeType",
            "description",
            "min_isv_svn",
            "minIsvSvn",
            "added_at",
            "addedAt",
            "added_by_proposal",
            "addedByProposal",
            "expiry_height",
            "expiryHeight",
            "revoked",
            "revoked_at",
            "revokedAt",
            "revoked_by_proposal",
            "revokedByProposal",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MeasurementHash,
            TeeType,
            Description,
            MinIsvSvn,
            AddedAt,
            AddedByProposal,
            ExpiryHeight,
            Revoked,
            RevokedAt,
            RevokedByProposal,
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
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "description" => Ok(GeneratedField::Description),
                            "minIsvSvn" | "min_isv_svn" => Ok(GeneratedField::MinIsvSvn),
                            "addedAt" | "added_at" => Ok(GeneratedField::AddedAt),
                            "addedByProposal" | "added_by_proposal" => Ok(GeneratedField::AddedByProposal),
                            "expiryHeight" | "expiry_height" => Ok(GeneratedField::ExpiryHeight),
                            "revoked" => Ok(GeneratedField::Revoked),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            "revokedByProposal" | "revoked_by_proposal" => Ok(GeneratedField::RevokedByProposal),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MeasurementRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MeasurementRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MeasurementRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurement_hash__ = None;
                let mut tee_type__ = None;
                let mut description__ = None;
                let mut min_isv_svn__ = None;
                let mut added_at__ = None;
                let mut added_by_proposal__ = None;
                let mut expiry_height__ = None;
                let mut revoked__ = None;
                let mut revoked_at__ = None;
                let mut revoked_by_proposal__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value::<TeeType>()? as i32);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinIsvSvn => {
                            if min_isv_svn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minIsvSvn"));
                            }
                            min_isv_svn__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AddedAt => {
                            if added_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addedAt"));
                            }
                            added_at__ = map_.next_value()?;
                        }
                        GeneratedField::AddedByProposal => {
                            if added_by_proposal__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addedByProposal"));
                            }
                            added_by_proposal__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiryHeight => {
                            if expiry_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryHeight"));
                            }
                            expiry_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Revoked => {
                            if revoked__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revoked"));
                            }
                            revoked__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = map_.next_value()?;
                        }
                        GeneratedField::RevokedByProposal => {
                            if revoked_by_proposal__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedByProposal"));
                            }
                            revoked_by_proposal__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MeasurementRecord {
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    tee_type: tee_type__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    min_isv_svn: min_isv_svn__.unwrap_or_default(),
                    added_at: added_at__,
                    added_by_proposal: added_by_proposal__.unwrap_or_default(),
                    expiry_height: expiry_height__.unwrap_or_default(),
                    revoked: revoked__.unwrap_or_default(),
                    revoked_at: revoked_at__,
                    revoked_by_proposal: revoked_by_proposal__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MeasurementRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgProposeMeasurement {
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
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if self.tee_type != 0 {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.min_isv_svn != 0 {
            len += 1;
        }
        if self.expiry_blocks != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgProposeMeasurement", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("measurementHash", pbjson::private::base64::encode(&self.measurement_hash).as_str())?;
        }
        if self.tee_type != 0 {
            let v = TeeType::try_from(self.tee_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tee_type)))?;
            struct_ser.serialize_field("teeType", &v)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.min_isv_svn != 0 {
            struct_ser.serialize_field("minIsvSvn", &self.min_isv_svn)?;
        }
        if self.expiry_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryBlocks", ToString::to_string(&self.expiry_blocks).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgProposeMeasurement {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "measurement_hash",
            "measurementHash",
            "tee_type",
            "teeType",
            "description",
            "min_isv_svn",
            "minIsvSvn",
            "expiry_blocks",
            "expiryBlocks",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            MeasurementHash,
            TeeType,
            Description,
            MinIsvSvn,
            ExpiryBlocks,
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
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "description" => Ok(GeneratedField::Description),
                            "minIsvSvn" | "min_isv_svn" => Ok(GeneratedField::MinIsvSvn),
                            "expiryBlocks" | "expiry_blocks" => Ok(GeneratedField::ExpiryBlocks),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgProposeMeasurement;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgProposeMeasurement")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgProposeMeasurement, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut measurement_hash__ = None;
                let mut tee_type__ = None;
                let mut description__ = None;
                let mut min_isv_svn__ = None;
                let mut expiry_blocks__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value::<TeeType>()? as i32);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinIsvSvn => {
                            if min_isv_svn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minIsvSvn"));
                            }
                            min_isv_svn__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiryBlocks => {
                            if expiry_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryBlocks"));
                            }
                            expiry_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgProposeMeasurement {
                    authority: authority__.unwrap_or_default(),
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    tee_type: tee_type__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    min_isv_svn: min_isv_svn__.unwrap_or_default(),
                    expiry_blocks: expiry_blocks__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgProposeMeasurement", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgProposeMeasurementResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgProposeMeasurementResponse", len)?;
        if !self.measurement_hash.is_empty() {
            struct_ser.serialize_field("measurementHash", &self.measurement_hash)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgProposeMeasurementResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurement_hash",
            "measurementHash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MeasurementHash,
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
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgProposeMeasurementResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgProposeMeasurementResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgProposeMeasurementResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurement_hash__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgProposeMeasurementResponse {
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgProposeMeasurementResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterEnclaveIdentity {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.tee_type != 0 {
            len += 1;
        }
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if !self.signer_hash.is_empty() {
            len += 1;
        }
        if !self.encryption_pub_key.is_empty() {
            len += 1;
        }
        if !self.signing_pub_key.is_empty() {
            len += 1;
        }
        if !self.attestation_quote.is_empty() {
            len += 1;
        }
        if !self.attestation_chain.is_empty() {
            len += 1;
        }
        if self.isv_prod_id != 0 {
            len += 1;
        }
        if self.isv_svn != 0 {
            len += 1;
        }
        if self.quote_version != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgRegisterEnclaveIdentity", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.tee_type != 0 {
            let v = TeeType::try_from(self.tee_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tee_type)))?;
            struct_ser.serialize_field("teeType", &v)?;
        }
        if !self.measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("measurementHash", pbjson::private::base64::encode(&self.measurement_hash).as_str())?;
        }
        if !self.signer_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signerHash", pbjson::private::base64::encode(&self.signer_hash).as_str())?;
        }
        if !self.encryption_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("encryptionPubKey", pbjson::private::base64::encode(&self.encryption_pub_key).as_str())?;
        }
        if !self.signing_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signingPubKey", pbjson::private::base64::encode(&self.signing_pub_key).as_str())?;
        }
        if !self.attestation_quote.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("attestationQuote", pbjson::private::base64::encode(&self.attestation_quote).as_str())?;
        }
        if !self.attestation_chain.is_empty() {
            struct_ser.serialize_field("attestationChain", &self.attestation_chain.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if self.isv_prod_id != 0 {
            struct_ser.serialize_field("isvProdId", &self.isv_prod_id)?;
        }
        if self.isv_svn != 0 {
            struct_ser.serialize_field("isvSvn", &self.isv_svn)?;
        }
        if self.quote_version != 0 {
            struct_ser.serialize_field("quoteVersion", &self.quote_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterEnclaveIdentity {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "tee_type",
            "teeType",
            "measurement_hash",
            "measurementHash",
            "signer_hash",
            "signerHash",
            "encryption_pub_key",
            "encryptionPubKey",
            "signing_pub_key",
            "signingPubKey",
            "attestation_quote",
            "attestationQuote",
            "attestation_chain",
            "attestationChain",
            "isv_prod_id",
            "isvProdId",
            "isv_svn",
            "isvSvn",
            "quote_version",
            "quoteVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            TeeType,
            MeasurementHash,
            SignerHash,
            EncryptionPubKey,
            SigningPubKey,
            AttestationQuote,
            AttestationChain,
            IsvProdId,
            IsvSvn,
            QuoteVersion,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "signerHash" | "signer_hash" => Ok(GeneratedField::SignerHash),
                            "encryptionPubKey" | "encryption_pub_key" => Ok(GeneratedField::EncryptionPubKey),
                            "signingPubKey" | "signing_pub_key" => Ok(GeneratedField::SigningPubKey),
                            "attestationQuote" | "attestation_quote" => Ok(GeneratedField::AttestationQuote),
                            "attestationChain" | "attestation_chain" => Ok(GeneratedField::AttestationChain),
                            "isvProdId" | "isv_prod_id" => Ok(GeneratedField::IsvProdId),
                            "isvSvn" | "isv_svn" => Ok(GeneratedField::IsvSvn),
                            "quoteVersion" | "quote_version" => Ok(GeneratedField::QuoteVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterEnclaveIdentity;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgRegisterEnclaveIdentity")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterEnclaveIdentity, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut tee_type__ = None;
                let mut measurement_hash__ = None;
                let mut signer_hash__ = None;
                let mut encryption_pub_key__ = None;
                let mut signing_pub_key__ = None;
                let mut attestation_quote__ = None;
                let mut attestation_chain__ = None;
                let mut isv_prod_id__ = None;
                let mut isv_svn__ = None;
                let mut quote_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value::<TeeType>()? as i32);
                        }
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SignerHash => {
                            if signer_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signerHash"));
                            }
                            signer_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EncryptionPubKey => {
                            if encryption_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptionPubKey"));
                            }
                            encryption_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SigningPubKey => {
                            if signing_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signingPubKey"));
                            }
                            signing_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationQuote => {
                            if attestation_quote__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationQuote"));
                            }
                            attestation_quote__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationChain => {
                            if attestation_chain__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationChain"));
                            }
                            attestation_chain__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::IsvProdId => {
                            if isv_prod_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isvProdId"));
                            }
                            isv_prod_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsvSvn => {
                            if isv_svn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isvSvn"));
                            }
                            isv_svn__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::QuoteVersion => {
                            if quote_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quoteVersion"));
                            }
                            quote_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRegisterEnclaveIdentity {
                    validator_address: validator_address__.unwrap_or_default(),
                    tee_type: tee_type__.unwrap_or_default(),
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    signer_hash: signer_hash__.unwrap_or_default(),
                    encryption_pub_key: encryption_pub_key__.unwrap_or_default(),
                    signing_pub_key: signing_pub_key__.unwrap_or_default(),
                    attestation_quote: attestation_quote__.unwrap_or_default(),
                    attestation_chain: attestation_chain__.unwrap_or_default(),
                    isv_prod_id: isv_prod_id__.unwrap_or_default(),
                    isv_svn: isv_svn__.unwrap_or_default(),
                    quote_version: quote_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgRegisterEnclaveIdentity", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterEnclaveIdentityResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.key_fingerprint.is_empty() {
            len += 1;
        }
        if self.expiry_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgRegisterEnclaveIdentityResponse", len)?;
        if !self.key_fingerprint.is_empty() {
            struct_ser.serialize_field("keyFingerprint", &self.key_fingerprint)?;
        }
        if self.expiry_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryHeight", ToString::to_string(&self.expiry_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterEnclaveIdentityResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key_fingerprint",
            "keyFingerprint",
            "expiry_height",
            "expiryHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            KeyFingerprint,
            ExpiryHeight,
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
                            "keyFingerprint" | "key_fingerprint" => Ok(GeneratedField::KeyFingerprint),
                            "expiryHeight" | "expiry_height" => Ok(GeneratedField::ExpiryHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterEnclaveIdentityResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgRegisterEnclaveIdentityResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterEnclaveIdentityResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key_fingerprint__ = None;
                let mut expiry_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::KeyFingerprint => {
                            if key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyFingerprint"));
                            }
                            key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiryHeight => {
                            if expiry_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryHeight"));
                            }
                            expiry_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRegisterEnclaveIdentityResponse {
                    key_fingerprint: key_fingerprint__.unwrap_or_default(),
                    expiry_height: expiry_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgRegisterEnclaveIdentityResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeMeasurement {
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
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgRevokeMeasurement", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("measurementHash", pbjson::private::base64::encode(&self.measurement_hash).as_str())?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeMeasurement {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "measurement_hash",
            "measurementHash",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            MeasurementHash,
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
                            "authority" => Ok(GeneratedField::Authority),
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
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
            type Value = MsgRevokeMeasurement;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgRevokeMeasurement")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeMeasurement, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut measurement_hash__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRevokeMeasurement {
                    authority: authority__.unwrap_or_default(),
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgRevokeMeasurement", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeMeasurementResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgRevokeMeasurementResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeMeasurementResponse {
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
            type Value = MsgRevokeMeasurementResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgRevokeMeasurementResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeMeasurementResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRevokeMeasurementResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgRevokeMeasurementResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRotateEnclaveIdentity {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.new_encryption_pub_key.is_empty() {
            len += 1;
        }
        if !self.new_signing_pub_key.is_empty() {
            len += 1;
        }
        if !self.new_attestation_quote.is_empty() {
            len += 1;
        }
        if !self.new_attestation_chain.is_empty() {
            len += 1;
        }
        if !self.new_measurement_hash.is_empty() {
            len += 1;
        }
        if self.new_isv_svn != 0 {
            len += 1;
        }
        if self.overlap_blocks != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgRotateEnclaveIdentity", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.new_encryption_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("newEncryptionPubKey", pbjson::private::base64::encode(&self.new_encryption_pub_key).as_str())?;
        }
        if !self.new_signing_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("newSigningPubKey", pbjson::private::base64::encode(&self.new_signing_pub_key).as_str())?;
        }
        if !self.new_attestation_quote.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("newAttestationQuote", pbjson::private::base64::encode(&self.new_attestation_quote).as_str())?;
        }
        if !self.new_attestation_chain.is_empty() {
            struct_ser.serialize_field("newAttestationChain", &self.new_attestation_chain.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.new_measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("newMeasurementHash", pbjson::private::base64::encode(&self.new_measurement_hash).as_str())?;
        }
        if self.new_isv_svn != 0 {
            struct_ser.serialize_field("newIsvSvn", &self.new_isv_svn)?;
        }
        if self.overlap_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapBlocks", ToString::to_string(&self.overlap_blocks).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRotateEnclaveIdentity {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "new_encryption_pub_key",
            "newEncryptionPubKey",
            "new_signing_pub_key",
            "newSigningPubKey",
            "new_attestation_quote",
            "newAttestationQuote",
            "new_attestation_chain",
            "newAttestationChain",
            "new_measurement_hash",
            "newMeasurementHash",
            "new_isv_svn",
            "newIsvSvn",
            "overlap_blocks",
            "overlapBlocks",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            NewEncryptionPubKey,
            NewSigningPubKey,
            NewAttestationQuote,
            NewAttestationChain,
            NewMeasurementHash,
            NewIsvSvn,
            OverlapBlocks,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "newEncryptionPubKey" | "new_encryption_pub_key" => Ok(GeneratedField::NewEncryptionPubKey),
                            "newSigningPubKey" | "new_signing_pub_key" => Ok(GeneratedField::NewSigningPubKey),
                            "newAttestationQuote" | "new_attestation_quote" => Ok(GeneratedField::NewAttestationQuote),
                            "newAttestationChain" | "new_attestation_chain" => Ok(GeneratedField::NewAttestationChain),
                            "newMeasurementHash" | "new_measurement_hash" => Ok(GeneratedField::NewMeasurementHash),
                            "newIsvSvn" | "new_isv_svn" => Ok(GeneratedField::NewIsvSvn),
                            "overlapBlocks" | "overlap_blocks" => Ok(GeneratedField::OverlapBlocks),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRotateEnclaveIdentity;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgRotateEnclaveIdentity")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRotateEnclaveIdentity, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut new_encryption_pub_key__ = None;
                let mut new_signing_pub_key__ = None;
                let mut new_attestation_quote__ = None;
                let mut new_attestation_chain__ = None;
                let mut new_measurement_hash__ = None;
                let mut new_isv_svn__ = None;
                let mut overlap_blocks__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewEncryptionPubKey => {
                            if new_encryption_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newEncryptionPubKey"));
                            }
                            new_encryption_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewSigningPubKey => {
                            if new_signing_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newSigningPubKey"));
                            }
                            new_signing_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewAttestationQuote => {
                            if new_attestation_quote__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newAttestationQuote"));
                            }
                            new_attestation_quote__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewAttestationChain => {
                            if new_attestation_chain__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newAttestationChain"));
                            }
                            new_attestation_chain__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::NewMeasurementHash => {
                            if new_measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newMeasurementHash"));
                            }
                            new_measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewIsvSvn => {
                            if new_isv_svn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newIsvSvn"));
                            }
                            new_isv_svn__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OverlapBlocks => {
                            if overlap_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapBlocks"));
                            }
                            overlap_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRotateEnclaveIdentity {
                    validator_address: validator_address__.unwrap_or_default(),
                    new_encryption_pub_key: new_encryption_pub_key__.unwrap_or_default(),
                    new_signing_pub_key: new_signing_pub_key__.unwrap_or_default(),
                    new_attestation_quote: new_attestation_quote__.unwrap_or_default(),
                    new_attestation_chain: new_attestation_chain__.unwrap_or_default(),
                    new_measurement_hash: new_measurement_hash__.unwrap_or_default(),
                    new_isv_svn: new_isv_svn__.unwrap_or_default(),
                    overlap_blocks: overlap_blocks__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgRotateEnclaveIdentity", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRotateEnclaveIdentityResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.new_key_fingerprint.is_empty() {
            len += 1;
        }
        if self.overlap_start_height != 0 {
            len += 1;
        }
        if self.overlap_end_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgRotateEnclaveIdentityResponse", len)?;
        if !self.new_key_fingerprint.is_empty() {
            struct_ser.serialize_field("newKeyFingerprint", &self.new_key_fingerprint)?;
        }
        if self.overlap_start_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapStartHeight", ToString::to_string(&self.overlap_start_height).as_str())?;
        }
        if self.overlap_end_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overlapEndHeight", ToString::to_string(&self.overlap_end_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRotateEnclaveIdentityResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "new_key_fingerprint",
            "newKeyFingerprint",
            "overlap_start_height",
            "overlapStartHeight",
            "overlap_end_height",
            "overlapEndHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            NewKeyFingerprint,
            OverlapStartHeight,
            OverlapEndHeight,
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
                            "newKeyFingerprint" | "new_key_fingerprint" => Ok(GeneratedField::NewKeyFingerprint),
                            "overlapStartHeight" | "overlap_start_height" => Ok(GeneratedField::OverlapStartHeight),
                            "overlapEndHeight" | "overlap_end_height" => Ok(GeneratedField::OverlapEndHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRotateEnclaveIdentityResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.MsgRotateEnclaveIdentityResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRotateEnclaveIdentityResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut new_key_fingerprint__ = None;
                let mut overlap_start_height__ = None;
                let mut overlap_end_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::NewKeyFingerprint => {
                            if new_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newKeyFingerprint"));
                            }
                            new_key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OverlapStartHeight => {
                            if overlap_start_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapStartHeight"));
                            }
                            overlap_start_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OverlapEndHeight => {
                            if overlap_end_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overlapEndHeight"));
                            }
                            overlap_end_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRotateEnclaveIdentityResponse {
                    new_key_fingerprint: new_key_fingerprint__.unwrap_or_default(),
                    overlap_start_height: overlap_start_height__.unwrap_or_default(),
                    overlap_end_height: overlap_end_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgRotateEnclaveIdentityResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.enclave.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.enclave.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.enclave.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.enclave.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if self.max_enclave_keys_per_validator != 0 {
            len += 1;
        }
        if self.default_expiry_blocks != 0 {
            len += 1;
        }
        if self.key_rotation_overlap_blocks != 0 {
            len += 1;
        }
        if self.min_quote_version != 0 {
            len += 1;
        }
        if !self.allowed_tee_types.is_empty() {
            len += 1;
        }
        if self.score_tolerance != 0 {
            len += 1;
        }
        if self.require_attestation_chain {
            len += 1;
        }
        if self.max_attestation_age != 0 {
            len += 1;
        }
        if self.enable_committee_mode {
            len += 1;
        }
        if self.committee_size != 0 {
            len += 1;
        }
        if self.committee_epoch_blocks != 0 {
            len += 1;
        }
        if self.enable_measurement_cleanup {
            len += 1;
        }
        if self.max_registrations_per_block != 0 {
            len += 1;
        }
        if self.registration_cooldown_blocks != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.Params", len)?;
        if self.max_enclave_keys_per_validator != 0 {
            struct_ser.serialize_field("maxEnclaveKeysPerValidator", &self.max_enclave_keys_per_validator)?;
        }
        if self.default_expiry_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("defaultExpiryBlocks", ToString::to_string(&self.default_expiry_blocks).as_str())?;
        }
        if self.key_rotation_overlap_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("keyRotationOverlapBlocks", ToString::to_string(&self.key_rotation_overlap_blocks).as_str())?;
        }
        if self.min_quote_version != 0 {
            struct_ser.serialize_field("minQuoteVersion", &self.min_quote_version)?;
        }
        if !self.allowed_tee_types.is_empty() {
            let v = self.allowed_tee_types.iter().cloned().map(|v| {
                TeeType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("allowedTeeTypes", &v)?;
        }
        if self.score_tolerance != 0 {
            struct_ser.serialize_field("scoreTolerance", &self.score_tolerance)?;
        }
        if self.require_attestation_chain {
            struct_ser.serialize_field("requireAttestationChain", &self.require_attestation_chain)?;
        }
        if self.max_attestation_age != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxAttestationAge", ToString::to_string(&self.max_attestation_age).as_str())?;
        }
        if self.enable_committee_mode {
            struct_ser.serialize_field("enableCommitteeMode", &self.enable_committee_mode)?;
        }
        if self.committee_size != 0 {
            struct_ser.serialize_field("committeeSize", &self.committee_size)?;
        }
        if self.committee_epoch_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("committeeEpochBlocks", ToString::to_string(&self.committee_epoch_blocks).as_str())?;
        }
        if self.enable_measurement_cleanup {
            struct_ser.serialize_field("enableMeasurementCleanup", &self.enable_measurement_cleanup)?;
        }
        if self.max_registrations_per_block != 0 {
            struct_ser.serialize_field("maxRegistrationsPerBlock", &self.max_registrations_per_block)?;
        }
        if self.registration_cooldown_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("registrationCooldownBlocks", ToString::to_string(&self.registration_cooldown_blocks).as_str())?;
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
            "max_enclave_keys_per_validator",
            "maxEnclaveKeysPerValidator",
            "default_expiry_blocks",
            "defaultExpiryBlocks",
            "key_rotation_overlap_blocks",
            "keyRotationOverlapBlocks",
            "min_quote_version",
            "minQuoteVersion",
            "allowed_tee_types",
            "allowedTeeTypes",
            "score_tolerance",
            "scoreTolerance",
            "require_attestation_chain",
            "requireAttestationChain",
            "max_attestation_age",
            "maxAttestationAge",
            "enable_committee_mode",
            "enableCommitteeMode",
            "committee_size",
            "committeeSize",
            "committee_epoch_blocks",
            "committeeEpochBlocks",
            "enable_measurement_cleanup",
            "enableMeasurementCleanup",
            "max_registrations_per_block",
            "maxRegistrationsPerBlock",
            "registration_cooldown_blocks",
            "registrationCooldownBlocks",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MaxEnclaveKeysPerValidator,
            DefaultExpiryBlocks,
            KeyRotationOverlapBlocks,
            MinQuoteVersion,
            AllowedTeeTypes,
            ScoreTolerance,
            RequireAttestationChain,
            MaxAttestationAge,
            EnableCommitteeMode,
            CommitteeSize,
            CommitteeEpochBlocks,
            EnableMeasurementCleanup,
            MaxRegistrationsPerBlock,
            RegistrationCooldownBlocks,
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
                            "maxEnclaveKeysPerValidator" | "max_enclave_keys_per_validator" => Ok(GeneratedField::MaxEnclaveKeysPerValidator),
                            "defaultExpiryBlocks" | "default_expiry_blocks" => Ok(GeneratedField::DefaultExpiryBlocks),
                            "keyRotationOverlapBlocks" | "key_rotation_overlap_blocks" => Ok(GeneratedField::KeyRotationOverlapBlocks),
                            "minQuoteVersion" | "min_quote_version" => Ok(GeneratedField::MinQuoteVersion),
                            "allowedTeeTypes" | "allowed_tee_types" => Ok(GeneratedField::AllowedTeeTypes),
                            "scoreTolerance" | "score_tolerance" => Ok(GeneratedField::ScoreTolerance),
                            "requireAttestationChain" | "require_attestation_chain" => Ok(GeneratedField::RequireAttestationChain),
                            "maxAttestationAge" | "max_attestation_age" => Ok(GeneratedField::MaxAttestationAge),
                            "enableCommitteeMode" | "enable_committee_mode" => Ok(GeneratedField::EnableCommitteeMode),
                            "committeeSize" | "committee_size" => Ok(GeneratedField::CommitteeSize),
                            "committeeEpochBlocks" | "committee_epoch_blocks" => Ok(GeneratedField::CommitteeEpochBlocks),
                            "enableMeasurementCleanup" | "enable_measurement_cleanup" => Ok(GeneratedField::EnableMeasurementCleanup),
                            "maxRegistrationsPerBlock" | "max_registrations_per_block" => Ok(GeneratedField::MaxRegistrationsPerBlock),
                            "registrationCooldownBlocks" | "registration_cooldown_blocks" => Ok(GeneratedField::RegistrationCooldownBlocks),
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
                formatter.write_str("struct virtengine.enclave.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut max_enclave_keys_per_validator__ = None;
                let mut default_expiry_blocks__ = None;
                let mut key_rotation_overlap_blocks__ = None;
                let mut min_quote_version__ = None;
                let mut allowed_tee_types__ = None;
                let mut score_tolerance__ = None;
                let mut require_attestation_chain__ = None;
                let mut max_attestation_age__ = None;
                let mut enable_committee_mode__ = None;
                let mut committee_size__ = None;
                let mut committee_epoch_blocks__ = None;
                let mut enable_measurement_cleanup__ = None;
                let mut max_registrations_per_block__ = None;
                let mut registration_cooldown_blocks__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MaxEnclaveKeysPerValidator => {
                            if max_enclave_keys_per_validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxEnclaveKeysPerValidator"));
                            }
                            max_enclave_keys_per_validator__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DefaultExpiryBlocks => {
                            if default_expiry_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("defaultExpiryBlocks"));
                            }
                            default_expiry_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::KeyRotationOverlapBlocks => {
                            if key_rotation_overlap_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyRotationOverlapBlocks"));
                            }
                            key_rotation_overlap_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinQuoteVersion => {
                            if min_quote_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minQuoteVersion"));
                            }
                            min_quote_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AllowedTeeTypes => {
                            if allowed_tee_types__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowedTeeTypes"));
                            }
                            allowed_tee_types__ = Some(map_.next_value::<Vec<TeeType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                        GeneratedField::ScoreTolerance => {
                            if score_tolerance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreTolerance"));
                            }
                            score_tolerance__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequireAttestationChain => {
                            if require_attestation_chain__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireAttestationChain"));
                            }
                            require_attestation_chain__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxAttestationAge => {
                            if max_attestation_age__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxAttestationAge"));
                            }
                            max_attestation_age__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EnableCommitteeMode => {
                            if enable_committee_mode__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enableCommitteeMode"));
                            }
                            enable_committee_mode__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CommitteeSize => {
                            if committee_size__.is_some() {
                                return Err(serde::de::Error::duplicate_field("committeeSize"));
                            }
                            committee_size__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CommitteeEpochBlocks => {
                            if committee_epoch_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("committeeEpochBlocks"));
                            }
                            committee_epoch_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EnableMeasurementCleanup => {
                            if enable_measurement_cleanup__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enableMeasurementCleanup"));
                            }
                            enable_measurement_cleanup__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxRegistrationsPerBlock => {
                            if max_registrations_per_block__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRegistrationsPerBlock"));
                            }
                            max_registrations_per_block__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RegistrationCooldownBlocks => {
                            if registration_cooldown_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registrationCooldownBlocks"));
                            }
                            registration_cooldown_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Params {
                    max_enclave_keys_per_validator: max_enclave_keys_per_validator__.unwrap_or_default(),
                    default_expiry_blocks: default_expiry_blocks__.unwrap_or_default(),
                    key_rotation_overlap_blocks: key_rotation_overlap_blocks__.unwrap_or_default(),
                    min_quote_version: min_quote_version__.unwrap_or_default(),
                    allowed_tee_types: allowed_tee_types__.unwrap_or_default(),
                    score_tolerance: score_tolerance__.unwrap_or_default(),
                    require_attestation_chain: require_attestation_chain__.unwrap_or_default(),
                    max_attestation_age: max_attestation_age__.unwrap_or_default(),
                    enable_committee_mode: enable_committee_mode__.unwrap_or_default(),
                    committee_size: committee_size__.unwrap_or_default(),
                    committee_epoch_blocks: committee_epoch_blocks__.unwrap_or_default(),
                    enable_measurement_cleanup: enable_measurement_cleanup__.unwrap_or_default(),
                    max_registrations_per_block: max_registrations_per_block__.unwrap_or_default(),
                    registration_cooldown_blocks: registration_cooldown_blocks__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryActiveValidatorEnclaveKeysRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysRequest", len)?;
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryActiveValidatorEnclaveKeysRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = QueryActiveValidatorEnclaveKeysRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryActiveValidatorEnclaveKeysRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryActiveValidatorEnclaveKeysRequest {
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryActiveValidatorEnclaveKeysResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.identities.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysResponse", len)?;
        if !self.identities.is_empty() {
            struct_ser.serialize_field("identities", &self.identities)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryActiveValidatorEnclaveKeysResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "identities",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Identities,
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
                            "identities" => Ok(GeneratedField::Identities),
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
            type Value = QueryActiveValidatorEnclaveKeysResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryActiveValidatorEnclaveKeysResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut identities__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Identities => {
                            if identities__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identities"));
                            }
                            identities__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryActiveValidatorEnclaveKeysResponse {
                    identities: identities__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryActiveValidatorEnclaveKeysResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAttestedResultRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.block_height != 0 {
            len += 1;
        }
        if !self.scope_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryAttestedResultRequest", len)?;
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAttestedResultRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "block_height",
            "blockHeight",
            "scope_id",
            "scopeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            BlockHeight,
            ScopeId,
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
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAttestedResultRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryAttestedResultRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAttestedResultRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut block_height__ = None;
                let mut scope_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAttestedResultRequest {
                    block_height: block_height__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryAttestedResultRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAttestedResultResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.result.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryAttestedResultResponse", len)?;
        if let Some(v) = self.result.as_ref() {
            struct_ser.serialize_field("result", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAttestedResultResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "result",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Result,
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
                            "result" => Ok(GeneratedField::Result),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAttestedResultResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryAttestedResultResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAttestedResultResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut result__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Result => {
                            if result__.is_some() {
                                return Err(serde::de::Error::duplicate_field("result"));
                            }
                            result__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAttestedResultResponse {
                    result: result__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryAttestedResultResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryCommitteeEnclaveKeysRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.committee_epoch != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryCommitteeEnclaveKeysRequest", len)?;
        if self.committee_epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("committeeEpoch", ToString::to_string(&self.committee_epoch).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryCommitteeEnclaveKeysRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "committee_epoch",
            "committeeEpoch",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CommitteeEpoch,
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
                            "committeeEpoch" | "committee_epoch" => Ok(GeneratedField::CommitteeEpoch),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryCommitteeEnclaveKeysRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryCommitteeEnclaveKeysRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryCommitteeEnclaveKeysRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut committee_epoch__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CommitteeEpoch => {
                            if committee_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("committeeEpoch"));
                            }
                            committee_epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryCommitteeEnclaveKeysRequest {
                    committee_epoch: committee_epoch__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryCommitteeEnclaveKeysRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryCommitteeEnclaveKeysResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.identities.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryCommitteeEnclaveKeysResponse", len)?;
        if !self.identities.is_empty() {
            struct_ser.serialize_field("identities", &self.identities)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryCommitteeEnclaveKeysResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "identities",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Identities,
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
                            "identities" => Ok(GeneratedField::Identities),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryCommitteeEnclaveKeysResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryCommitteeEnclaveKeysResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryCommitteeEnclaveKeysResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut identities__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Identities => {
                            if identities__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identities"));
                            }
                            identities__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryCommitteeEnclaveKeysResponse {
                    identities: identities__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryCommitteeEnclaveKeysResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryEnclaveIdentityRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryEnclaveIdentityRequest", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryEnclaveIdentityRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryEnclaveIdentityRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryEnclaveIdentityRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryEnclaveIdentityRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryEnclaveIdentityRequest {
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryEnclaveIdentityRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryEnclaveIdentityResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.identity.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryEnclaveIdentityResponse", len)?;
        if let Some(v) = self.identity.as_ref() {
            struct_ser.serialize_field("identity", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryEnclaveIdentityResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "identity",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Identity,
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
                            "identity" => Ok(GeneratedField::Identity),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryEnclaveIdentityResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryEnclaveIdentityResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryEnclaveIdentityResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut identity__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Identity => {
                            if identity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identity"));
                            }
                            identity__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryEnclaveIdentityResponse {
                    identity: identity__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryEnclaveIdentityResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryKeyRotationRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryKeyRotationRequest", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryKeyRotationRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryKeyRotationRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryKeyRotationRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryKeyRotationRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryKeyRotationRequest {
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryKeyRotationRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryKeyRotationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.rotation.is_some() {
            len += 1;
        }
        if self.has_active_rotation {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryKeyRotationResponse", len)?;
        if let Some(v) = self.rotation.as_ref() {
            struct_ser.serialize_field("rotation", v)?;
        }
        if self.has_active_rotation {
            struct_ser.serialize_field("hasActiveRotation", &self.has_active_rotation)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryKeyRotationResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "rotation",
            "has_active_rotation",
            "hasActiveRotation",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Rotation,
            HasActiveRotation,
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
                            "rotation" => Ok(GeneratedField::Rotation),
                            "hasActiveRotation" | "has_active_rotation" => Ok(GeneratedField::HasActiveRotation),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryKeyRotationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryKeyRotationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryKeyRotationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut rotation__ = None;
                let mut has_active_rotation__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Rotation => {
                            if rotation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rotation"));
                            }
                            rotation__ = map_.next_value()?;
                        }
                        GeneratedField::HasActiveRotation => {
                            if has_active_rotation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hasActiveRotation"));
                            }
                            has_active_rotation__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryKeyRotationResponse {
                    rotation: rotation__,
                    has_active_rotation: has_active_rotation__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryKeyRotationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMeasurementAllowlistRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.tee_type.is_empty() {
            len += 1;
        }
        if self.include_revoked {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryMeasurementAllowlistRequest", len)?;
        if !self.tee_type.is_empty() {
            struct_ser.serialize_field("teeType", &self.tee_type)?;
        }
        if self.include_revoked {
            struct_ser.serialize_field("includeRevoked", &self.include_revoked)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMeasurementAllowlistRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "tee_type",
            "teeType",
            "include_revoked",
            "includeRevoked",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TeeType,
            IncludeRevoked,
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
                            "teeType" | "tee_type" => Ok(GeneratedField::TeeType),
                            "includeRevoked" | "include_revoked" => Ok(GeneratedField::IncludeRevoked),
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
            type Value = QueryMeasurementAllowlistRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryMeasurementAllowlistRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMeasurementAllowlistRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut tee_type__ = None;
                let mut include_revoked__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TeeType => {
                            if tee_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("teeType"));
                            }
                            tee_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IncludeRevoked => {
                            if include_revoked__.is_some() {
                                return Err(serde::de::Error::duplicate_field("includeRevoked"));
                            }
                            include_revoked__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryMeasurementAllowlistRequest {
                    tee_type: tee_type__.unwrap_or_default(),
                    include_revoked: include_revoked__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryMeasurementAllowlistRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMeasurementAllowlistResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.measurements.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryMeasurementAllowlistResponse", len)?;
        if !self.measurements.is_empty() {
            struct_ser.serialize_field("measurements", &self.measurements)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMeasurementAllowlistResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurements",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Measurements,
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
                            "measurements" => Ok(GeneratedField::Measurements),
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
            type Value = QueryMeasurementAllowlistResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryMeasurementAllowlistResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMeasurementAllowlistResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurements__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Measurements => {
                            if measurements__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurements"));
                            }
                            measurements__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryMeasurementAllowlistResponse {
                    measurements: measurements__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryMeasurementAllowlistResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMeasurementRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryMeasurementRequest", len)?;
        if !self.measurement_hash.is_empty() {
            struct_ser.serialize_field("measurementHash", &self.measurement_hash)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMeasurementRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurement_hash",
            "measurementHash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MeasurementHash,
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
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryMeasurementRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryMeasurementRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMeasurementRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurement_hash__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryMeasurementRequest {
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryMeasurementRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMeasurementResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.measurement.is_some() {
            len += 1;
        }
        if self.is_allowed {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryMeasurementResponse", len)?;
        if let Some(v) = self.measurement.as_ref() {
            struct_ser.serialize_field("measurement", v)?;
        }
        if self.is_allowed {
            struct_ser.serialize_field("isAllowed", &self.is_allowed)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMeasurementResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "measurement",
            "is_allowed",
            "isAllowed",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Measurement,
            IsAllowed,
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
                            "measurement" => Ok(GeneratedField::Measurement),
                            "isAllowed" | "is_allowed" => Ok(GeneratedField::IsAllowed),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryMeasurementResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryMeasurementResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMeasurementResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut measurement__ = None;
                let mut is_allowed__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Measurement => {
                            if measurement__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurement"));
                            }
                            measurement__ = map_.next_value()?;
                        }
                        GeneratedField::IsAllowed => {
                            if is_allowed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isAllowed"));
                            }
                            is_allowed__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryMeasurementResponse {
                    measurement: measurement__,
                    is_allowed: is_allowed__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryMeasurementResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.enclave.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.enclave.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidKeySetRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.for_block_height != 0 {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryValidKeySetRequest", len)?;
        if self.for_block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("forBlockHeight", ToString::to_string(&self.for_block_height).as_str())?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidKeySetRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "for_block_height",
            "forBlockHeight",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ForBlockHeight,
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
                            "forBlockHeight" | "for_block_height" => Ok(GeneratedField::ForBlockHeight),
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
            type Value = QueryValidKeySetRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryValidKeySetRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidKeySetRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut for_block_height__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ForBlockHeight => {
                            if for_block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("forBlockHeight"));
                            }
                            for_block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryValidKeySetRequest {
                    for_block_height: for_block_height__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryValidKeySetRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidKeySetResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_keys.is_empty() {
            len += 1;
        }
        if self.total_count != 0 {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.QueryValidKeySetResponse", len)?;
        if !self.validator_keys.is_empty() {
            struct_ser.serialize_field("validatorKeys", &self.validator_keys)?;
        }
        if self.total_count != 0 {
            struct_ser.serialize_field("totalCount", &self.total_count)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidKeySetResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_keys",
            "validatorKeys",
            "total_count",
            "totalCount",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorKeys,
            TotalCount,
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
                            "validatorKeys" | "validator_keys" => Ok(GeneratedField::ValidatorKeys),
                            "totalCount" | "total_count" => Ok(GeneratedField::TotalCount),
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
            type Value = QueryValidKeySetResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.QueryValidKeySetResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidKeySetResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_keys__ = None;
                let mut total_count__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorKeys => {
                            if validator_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorKeys"));
                            }
                            validator_keys__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalCount => {
                            if total_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalCount"));
                            }
                            total_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryValidKeySetResponse {
                    validator_keys: validator_keys__.unwrap_or_default(),
                    total_count: total_count__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.QueryValidKeySetResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for TeeType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "TEE_TYPE_UNSPECIFIED",
            Self::Sgx => "TEE_TYPE_SGX",
            Self::SevSnp => "TEE_TYPE_SEV_SNP",
            Self::Nitro => "TEE_TYPE_NITRO",
            Self::Trustzone => "TEE_TYPE_TRUSTZONE",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for TeeType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "TEE_TYPE_UNSPECIFIED",
            "TEE_TYPE_SGX",
            "TEE_TYPE_SEV_SNP",
            "TEE_TYPE_NITRO",
            "TEE_TYPE_TRUSTZONE",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = TeeType;

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
                    "TEE_TYPE_UNSPECIFIED" => Ok(TeeType::Unspecified),
                    "TEE_TYPE_SGX" => Ok(TeeType::Sgx),
                    "TEE_TYPE_SEV_SNP" => Ok(TeeType::SevSnp),
                    "TEE_TYPE_NITRO" => Ok(TeeType::Nitro),
                    "TEE_TYPE_TRUSTZONE" => Ok(TeeType::Trustzone),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorKeyInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.encryption_key_id.is_empty() {
            len += 1;
        }
        if !self.encryption_pub_key.is_empty() {
            len += 1;
        }
        if !self.measurement_hash.is_empty() {
            len += 1;
        }
        if self.expiry_height != 0 {
            len += 1;
        }
        if self.is_in_rotation {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.enclave.v1.ValidatorKeyInfo", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.encryption_key_id.is_empty() {
            struct_ser.serialize_field("encryptionKeyId", &self.encryption_key_id)?;
        }
        if !self.encryption_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("encryptionPubKey", pbjson::private::base64::encode(&self.encryption_pub_key).as_str())?;
        }
        if !self.measurement_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("measurementHash", pbjson::private::base64::encode(&self.measurement_hash).as_str())?;
        }
        if self.expiry_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryHeight", ToString::to_string(&self.expiry_height).as_str())?;
        }
        if self.is_in_rotation {
            struct_ser.serialize_field("isInRotation", &self.is_in_rotation)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValidatorKeyInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "encryption_key_id",
            "encryptionKeyId",
            "encryption_pub_key",
            "encryptionPubKey",
            "measurement_hash",
            "measurementHash",
            "expiry_height",
            "expiryHeight",
            "is_in_rotation",
            "isInRotation",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            EncryptionKeyId,
            EncryptionPubKey,
            MeasurementHash,
            ExpiryHeight,
            IsInRotation,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "encryptionKeyId" | "encryption_key_id" => Ok(GeneratedField::EncryptionKeyId),
                            "encryptionPubKey" | "encryption_pub_key" => Ok(GeneratedField::EncryptionPubKey),
                            "measurementHash" | "measurement_hash" => Ok(GeneratedField::MeasurementHash),
                            "expiryHeight" | "expiry_height" => Ok(GeneratedField::ExpiryHeight),
                            "isInRotation" | "is_in_rotation" => Ok(GeneratedField::IsInRotation),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValidatorKeyInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.enclave.v1.ValidatorKeyInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValidatorKeyInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut encryption_key_id__ = None;
                let mut encryption_pub_key__ = None;
                let mut measurement_hash__ = None;
                let mut expiry_height__ = None;
                let mut is_in_rotation__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptionKeyId => {
                            if encryption_key_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptionKeyId"));
                            }
                            encryption_key_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptionPubKey => {
                            if encryption_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptionPubKey"));
                            }
                            encryption_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MeasurementHash => {
                            if measurement_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measurementHash"));
                            }
                            measurement_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiryHeight => {
                            if expiry_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryHeight"));
                            }
                            expiry_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsInRotation => {
                            if is_in_rotation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isInRotation"));
                            }
                            is_in_rotation__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ValidatorKeyInfo {
                    validator_address: validator_address__.unwrap_or_default(),
                    encryption_key_id: encryption_key_id__.unwrap_or_default(),
                    encryption_pub_key: encryption_pub_key__.unwrap_or_default(),
                    measurement_hash: measurement_hash__.unwrap_or_default(),
                    expiry_height: expiry_height__.unwrap_or_default(),
                    is_in_rotation: is_in_rotation__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.enclave.v1.ValidatorKeyInfo", FIELDS, GeneratedVisitor)
    }
}

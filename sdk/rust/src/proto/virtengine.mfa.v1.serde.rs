// @generated
impl serde::Serialize for AuthorizationSession {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.session_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.transaction_type != 0 {
            len += 1;
        }
        if !self.verified_factors.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if self.used_at != 0 {
            len += 1;
        }
        if self.is_single_use {
            len += 1;
        }
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.AuthorizationSession", len)?;
        if !self.session_id.is_empty() {
            struct_ser.serialize_field("sessionId", &self.session_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.transaction_type != 0 {
            let v = SensitiveTransactionType::try_from(self.transaction_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.transaction_type)))?;
            struct_ser.serialize_field("transactionType", &v)?;
        }
        if !self.verified_factors.is_empty() {
            let v = self.verified_factors.iter().cloned().map(|v| {
                FactorType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("verifiedFactors", &v)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if self.used_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("usedAt", ToString::to_string(&self.used_at).as_str())?;
        }
        if self.is_single_use {
            struct_ser.serialize_field("isSingleUse", &self.is_single_use)?;
        }
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AuthorizationSession {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "session_id",
            "sessionId",
            "account_address",
            "accountAddress",
            "transaction_type",
            "transactionType",
            "verified_factors",
            "verifiedFactors",
            "created_at",
            "createdAt",
            "expires_at",
            "expiresAt",
            "used_at",
            "usedAt",
            "is_single_use",
            "isSingleUse",
            "device_fingerprint",
            "deviceFingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SessionId,
            AccountAddress,
            TransactionType,
            VerifiedFactors,
            CreatedAt,
            ExpiresAt,
            UsedAt,
            IsSingleUse,
            DeviceFingerprint,
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
                            "sessionId" | "session_id" => Ok(GeneratedField::SessionId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            "verifiedFactors" | "verified_factors" => Ok(GeneratedField::VerifiedFactors),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "usedAt" | "used_at" => Ok(GeneratedField::UsedAt),
                            "isSingleUse" | "is_single_use" => Ok(GeneratedField::IsSingleUse),
                            "deviceFingerprint" | "device_fingerprint" => Ok(GeneratedField::DeviceFingerprint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AuthorizationSession;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.AuthorizationSession")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AuthorizationSession, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut session_id__ = None;
                let mut account_address__ = None;
                let mut transaction_type__ = None;
                let mut verified_factors__ = None;
                let mut created_at__ = None;
                let mut expires_at__ = None;
                let mut used_at__ = None;
                let mut is_single_use__ = None;
                let mut device_fingerprint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SessionId => {
                            if session_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionId"));
                            }
                            session_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value::<SensitiveTransactionType>()? as i32);
                        }
                        GeneratedField::VerifiedFactors => {
                            if verified_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedFactors"));
                            }
                            verified_factors__ = Some(map_.next_value::<Vec<FactorType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UsedAt => {
                            if used_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usedAt"));
                            }
                            used_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsSingleUse => {
                            if is_single_use__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isSingleUse"));
                            }
                            is_single_use__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(AuthorizationSession {
                    session_id: session_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    transaction_type: transaction_type__.unwrap_or_default(),
                    verified_factors: verified_factors__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    used_at: used_at__.unwrap_or_default(),
                    is_single_use: is_single_use__.unwrap_or_default(),
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.AuthorizationSession", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Challenge {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.factor_type != 0 {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        if self.transaction_type != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.challenge_data.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if self.verified_at != 0 {
            len += 1;
        }
        if self.attempt_count != 0 {
            len += 1;
        }
        if self.max_attempts != 0 {
            len += 1;
        }
        if !self.nonce.is_empty() {
            len += 1;
        }
        if !self.session_id.is_empty() {
            len += 1;
        }
        if self.metadata.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.Challenge", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.factor_type != 0 {
            let v = FactorType::try_from(self.factor_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type)))?;
            struct_ser.serialize_field("factorType", &v)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        if self.transaction_type != 0 {
            let v = SensitiveTransactionType::try_from(self.transaction_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.transaction_type)))?;
            struct_ser.serialize_field("transactionType", &v)?;
        }
        if self.status != 0 {
            let v = ChallengeStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.challenge_data.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challengeData", pbjson::private::base64::encode(&self.challenge_data).as_str())?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if self.verified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("verifiedAt", ToString::to_string(&self.verified_at).as_str())?;
        }
        if self.attempt_count != 0 {
            struct_ser.serialize_field("attemptCount", &self.attempt_count)?;
        }
        if self.max_attempts != 0 {
            struct_ser.serialize_field("maxAttempts", &self.max_attempts)?;
        }
        if !self.nonce.is_empty() {
            struct_ser.serialize_field("nonce", &self.nonce)?;
        }
        if !self.session_id.is_empty() {
            struct_ser.serialize_field("sessionId", &self.session_id)?;
        }
        if let Some(v) = self.metadata.as_ref() {
            struct_ser.serialize_field("metadata", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Challenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "account_address",
            "accountAddress",
            "factor_type",
            "factorType",
            "factor_id",
            "factorId",
            "transaction_type",
            "transactionType",
            "status",
            "challenge_data",
            "challengeData",
            "created_at",
            "createdAt",
            "expires_at",
            "expiresAt",
            "verified_at",
            "verifiedAt",
            "attempt_count",
            "attemptCount",
            "max_attempts",
            "maxAttempts",
            "nonce",
            "session_id",
            "sessionId",
            "metadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            AccountAddress,
            FactorType,
            FactorId,
            TransactionType,
            Status,
            ChallengeData,
            CreatedAt,
            ExpiresAt,
            VerifiedAt,
            AttemptCount,
            MaxAttempts,
            Nonce,
            SessionId,
            Metadata,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            "status" => Ok(GeneratedField::Status),
                            "challengeData" | "challenge_data" => Ok(GeneratedField::ChallengeData),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "verifiedAt" | "verified_at" => Ok(GeneratedField::VerifiedAt),
                            "attemptCount" | "attempt_count" => Ok(GeneratedField::AttemptCount),
                            "maxAttempts" | "max_attempts" => Ok(GeneratedField::MaxAttempts),
                            "nonce" => Ok(GeneratedField::Nonce),
                            "sessionId" | "session_id" => Ok(GeneratedField::SessionId),
                            "metadata" => Ok(GeneratedField::Metadata),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Challenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.Challenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Challenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut account_address__ = None;
                let mut factor_type__ = None;
                let mut factor_id__ = None;
                let mut transaction_type__ = None;
                let mut status__ = None;
                let mut challenge_data__ = None;
                let mut created_at__ = None;
                let mut expires_at__ = None;
                let mut verified_at__ = None;
                let mut attempt_count__ = None;
                let mut max_attempts__ = None;
                let mut nonce__ = None;
                let mut session_id__ = None;
                let mut metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value::<SensitiveTransactionType>()? as i32);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ChallengeStatus>()? as i32);
                        }
                        GeneratedField::ChallengeData => {
                            if challenge_data__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeData"));
                            }
                            challenge_data__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VerifiedAt => {
                            if verified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedAt"));
                            }
                            verified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttemptCount => {
                            if attempt_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attemptCount"));
                            }
                            attempt_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxAttempts => {
                            if max_attempts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxAttempts"));
                            }
                            max_attempts__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Nonce => {
                            if nonce__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nonce"));
                            }
                            nonce__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SessionId => {
                            if session_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionId"));
                            }
                            session_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Challenge {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                    transaction_type: transaction_type__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    challenge_data: challenge_data__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    verified_at: verified_at__.unwrap_or_default(),
                    attempt_count: attempt_count__.unwrap_or_default(),
                    max_attempts: max_attempts__.unwrap_or_default(),
                    nonce: nonce__.unwrap_or_default(),
                    session_id: session_id__.unwrap_or_default(),
                    metadata: metadata__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.Challenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ChallengeMetadata {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.fido2_challenge.is_some() {
            len += 1;
        }
        if self.otp_info.is_some() {
            len += 1;
        }
        if self.client_info.is_some() {
            len += 1;
        }
        if self.hardware_key_challenge.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.ChallengeMetadata", len)?;
        if let Some(v) = self.fido2_challenge.as_ref() {
            struct_ser.serialize_field("fido2Challenge", v)?;
        }
        if let Some(v) = self.otp_info.as_ref() {
            struct_ser.serialize_field("otpInfo", v)?;
        }
        if let Some(v) = self.client_info.as_ref() {
            struct_ser.serialize_field("clientInfo", v)?;
        }
        if let Some(v) = self.hardware_key_challenge.as_ref() {
            struct_ser.serialize_field("hardwareKeyChallenge", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ChallengeMetadata {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "fido2_challenge",
            "fido2Challenge",
            "otp_info",
            "otpInfo",
            "client_info",
            "clientInfo",
            "hardware_key_challenge",
            "hardwareKeyChallenge",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Fido2Challenge,
            OtpInfo,
            ClientInfo,
            HardwareKeyChallenge,
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
                            "fido2Challenge" | "fido2_challenge" => Ok(GeneratedField::Fido2Challenge),
                            "otpInfo" | "otp_info" => Ok(GeneratedField::OtpInfo),
                            "clientInfo" | "client_info" => Ok(GeneratedField::ClientInfo),
                            "hardwareKeyChallenge" | "hardware_key_challenge" => Ok(GeneratedField::HardwareKeyChallenge),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ChallengeMetadata;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.ChallengeMetadata")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ChallengeMetadata, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut fido2_challenge__ = None;
                let mut otp_info__ = None;
                let mut client_info__ = None;
                let mut hardware_key_challenge__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Fido2Challenge => {
                            if fido2_challenge__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fido2Challenge"));
                            }
                            fido2_challenge__ = map_.next_value()?;
                        }
                        GeneratedField::OtpInfo => {
                            if otp_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("otpInfo"));
                            }
                            otp_info__ = map_.next_value()?;
                        }
                        GeneratedField::ClientInfo => {
                            if client_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientInfo"));
                            }
                            client_info__ = map_.next_value()?;
                        }
                        GeneratedField::HardwareKeyChallenge => {
                            if hardware_key_challenge__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hardwareKeyChallenge"));
                            }
                            hardware_key_challenge__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ChallengeMetadata {
                    fido2_challenge: fido2_challenge__,
                    otp_info: otp_info__,
                    client_info: client_info__,
                    hardware_key_challenge: hardware_key_challenge__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.ChallengeMetadata", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ChallengeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if self.factor_type != 0 {
            len += 1;
        }
        if !self.response_data.is_empty() {
            len += 1;
        }
        if self.client_info.is_some() {
            len += 1;
        }
        if self.timestamp != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.ChallengeResponse", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if self.factor_type != 0 {
            let v = FactorType::try_from(self.factor_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type)))?;
            struct_ser.serialize_field("factorType", &v)?;
        }
        if !self.response_data.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("responseData", pbjson::private::base64::encode(&self.response_data).as_str())?;
        }
        if let Some(v) = self.client_info.as_ref() {
            struct_ser.serialize_field("clientInfo", v)?;
        }
        if self.timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("timestamp", ToString::to_string(&self.timestamp).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ChallengeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "factor_type",
            "factorType",
            "response_data",
            "responseData",
            "client_info",
            "clientInfo",
            "timestamp",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            FactorType,
            ResponseData,
            ClientInfo,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "responseData" | "response_data" => Ok(GeneratedField::ResponseData),
                            "clientInfo" | "client_info" => Ok(GeneratedField::ClientInfo),
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
            type Value = ChallengeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.ChallengeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ChallengeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut factor_type__ = None;
                let mut response_data__ = None;
                let mut client_info__ = None;
                let mut timestamp__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::ResponseData => {
                            if response_data__.is_some() {
                                return Err(serde::de::Error::duplicate_field("responseData"));
                            }
                            response_data__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ClientInfo => {
                            if client_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientInfo"));
                            }
                            client_info__ = map_.next_value()?;
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ChallengeResponse {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    response_data: response_data__.unwrap_or_default(),
                    client_info: client_info__,
                    timestamp: timestamp__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.ChallengeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ChallengeStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "CHALLENGE_STATUS_UNSPECIFIED",
            Self::Pending => "CHALLENGE_STATUS_PENDING",
            Self::Verified => "CHALLENGE_STATUS_VERIFIED",
            Self::Failed => "CHALLENGE_STATUS_FAILED",
            Self::Expired => "CHALLENGE_STATUS_EXPIRED",
            Self::Cancelled => "CHALLENGE_STATUS_CANCELLED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ChallengeStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "CHALLENGE_STATUS_UNSPECIFIED",
            "CHALLENGE_STATUS_PENDING",
            "CHALLENGE_STATUS_VERIFIED",
            "CHALLENGE_STATUS_FAILED",
            "CHALLENGE_STATUS_EXPIRED",
            "CHALLENGE_STATUS_CANCELLED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ChallengeStatus;

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
                    "CHALLENGE_STATUS_UNSPECIFIED" => Ok(ChallengeStatus::Unspecified),
                    "CHALLENGE_STATUS_PENDING" => Ok(ChallengeStatus::Pending),
                    "CHALLENGE_STATUS_VERIFIED" => Ok(ChallengeStatus::Verified),
                    "CHALLENGE_STATUS_FAILED" => Ok(ChallengeStatus::Failed),
                    "CHALLENGE_STATUS_EXPIRED" => Ok(ChallengeStatus::Expired),
                    "CHALLENGE_STATUS_CANCELLED" => Ok(ChallengeStatus::Cancelled),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ClientInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        if !self.ip_hash.is_empty() {
            len += 1;
        }
        if !self.user_agent.is_empty() {
            len += 1;
        }
        if self.requested_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.ClientInfo", len)?;
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        if !self.ip_hash.is_empty() {
            struct_ser.serialize_field("ipHash", &self.ip_hash)?;
        }
        if !self.user_agent.is_empty() {
            struct_ser.serialize_field("userAgent", &self.user_agent)?;
        }
        if self.requested_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("requestedAt", ToString::to_string(&self.requested_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ClientInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "device_fingerprint",
            "deviceFingerprint",
            "ip_hash",
            "ipHash",
            "user_agent",
            "userAgent",
            "requested_at",
            "requestedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DeviceFingerprint,
            IpHash,
            UserAgent,
            RequestedAt,
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
                            "deviceFingerprint" | "device_fingerprint" => Ok(GeneratedField::DeviceFingerprint),
                            "ipHash" | "ip_hash" => Ok(GeneratedField::IpHash),
                            "userAgent" | "user_agent" => Ok(GeneratedField::UserAgent),
                            "requestedAt" | "requested_at" => Ok(GeneratedField::RequestedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ClientInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.ClientInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ClientInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut device_fingerprint__ = None;
                let mut ip_hash__ = None;
                let mut user_agent__ = None;
                let mut requested_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IpHash => {
                            if ip_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ipHash"));
                            }
                            ip_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UserAgent => {
                            if user_agent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userAgent"));
                            }
                            user_agent__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequestedAt => {
                            if requested_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requestedAt"));
                            }
                            requested_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ClientInfo {
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                    ip_hash: ip_hash__.unwrap_or_default(),
                    user_agent: user_agent__.unwrap_or_default(),
                    requested_at: requested_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.ClientInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for DeviceInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.fingerprint.is_empty() {
            len += 1;
        }
        if !self.user_agent.is_empty() {
            len += 1;
        }
        if self.first_seen_at != 0 {
            len += 1;
        }
        if self.last_seen_at != 0 {
            len += 1;
        }
        if !self.ip_hash.is_empty() {
            len += 1;
        }
        if self.trust_expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.DeviceInfo", len)?;
        if !self.fingerprint.is_empty() {
            struct_ser.serialize_field("fingerprint", &self.fingerprint)?;
        }
        if !self.user_agent.is_empty() {
            struct_ser.serialize_field("userAgent", &self.user_agent)?;
        }
        if self.first_seen_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("firstSeenAt", ToString::to_string(&self.first_seen_at).as_str())?;
        }
        if self.last_seen_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastSeenAt", ToString::to_string(&self.last_seen_at).as_str())?;
        }
        if !self.ip_hash.is_empty() {
            struct_ser.serialize_field("ipHash", &self.ip_hash)?;
        }
        if self.trust_expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("trustExpiresAt", ToString::to_string(&self.trust_expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DeviceInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "fingerprint",
            "user_agent",
            "userAgent",
            "first_seen_at",
            "firstSeenAt",
            "last_seen_at",
            "lastSeenAt",
            "ip_hash",
            "ipHash",
            "trust_expires_at",
            "trustExpiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Fingerprint,
            UserAgent,
            FirstSeenAt,
            LastSeenAt,
            IpHash,
            TrustExpiresAt,
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
                            "fingerprint" => Ok(GeneratedField::Fingerprint),
                            "userAgent" | "user_agent" => Ok(GeneratedField::UserAgent),
                            "firstSeenAt" | "first_seen_at" => Ok(GeneratedField::FirstSeenAt),
                            "lastSeenAt" | "last_seen_at" => Ok(GeneratedField::LastSeenAt),
                            "ipHash" | "ip_hash" => Ok(GeneratedField::IpHash),
                            "trustExpiresAt" | "trust_expires_at" => Ok(GeneratedField::TrustExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DeviceInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.DeviceInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DeviceInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut fingerprint__ = None;
                let mut user_agent__ = None;
                let mut first_seen_at__ = None;
                let mut last_seen_at__ = None;
                let mut ip_hash__ = None;
                let mut trust_expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Fingerprint => {
                            if fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fingerprint"));
                            }
                            fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UserAgent => {
                            if user_agent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userAgent"));
                            }
                            user_agent__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FirstSeenAt => {
                            if first_seen_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("firstSeenAt"));
                            }
                            first_seen_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastSeenAt => {
                            if last_seen_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastSeenAt"));
                            }
                            last_seen_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IpHash => {
                            if ip_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ipHash"));
                            }
                            ip_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TrustExpiresAt => {
                            if trust_expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustExpiresAt"));
                            }
                            trust_expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(DeviceInfo {
                    fingerprint: fingerprint__.unwrap_or_default(),
                    user_agent: user_agent__.unwrap_or_default(),
                    first_seen_at: first_seen_at__.unwrap_or_default(),
                    last_seen_at: last_seen_at__.unwrap_or_default(),
                    ip_hash: ip_hash__.unwrap_or_default(),
                    trust_expires_at: trust_expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.DeviceInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventChallengeVerified {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.factor_type.is_empty() {
            len += 1;
        }
        if !self.transaction_type.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.EventChallengeVerified", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.factor_type.is_empty() {
            struct_ser.serialize_field("factorType", &self.factor_type)?;
        }
        if !self.transaction_type.is_empty() {
            struct_ser.serialize_field("transactionType", &self.transaction_type)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventChallengeVerified {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "challenge_id",
            "challengeId",
            "factor_type",
            "factorType",
            "transaction_type",
            "transactionType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ChallengeId,
            FactorType,
            TransactionType,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventChallengeVerified;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.EventChallengeVerified")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventChallengeVerified, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut challenge_id__ = None;
                let mut factor_type__ = None;
                let mut transaction_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventChallengeVerified {
                    account_address: account_address__.unwrap_or_default(),
                    challenge_id: challenge_id__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    transaction_type: transaction_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.EventChallengeVerified", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventFactorEnrolled {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.factor_type.is_empty() {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.EventFactorEnrolled", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.factor_type.is_empty() {
            struct_ser.serialize_field("factorType", &self.factor_type)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventFactorEnrolled {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "factor_type",
            "factorType",
            "factor_id",
            "factorId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            FactorType,
            FactorId,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventFactorEnrolled;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.EventFactorEnrolled")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventFactorEnrolled, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut factor_type__ = None;
                let mut factor_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventFactorEnrolled {
                    account_address: account_address__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.EventFactorEnrolled", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventFactorRevoked {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.factor_type.is_empty() {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.EventFactorRevoked", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.factor_type.is_empty() {
            struct_ser.serialize_field("factorType", &self.factor_type)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventFactorRevoked {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "factor_type",
            "factorType",
            "factor_id",
            "factorId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            FactorType,
            FactorId,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventFactorRevoked;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.EventFactorRevoked")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventFactorRevoked, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut factor_type__ = None;
                let mut factor_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventFactorRevoked {
                    account_address: account_address__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.EventFactorRevoked", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventMfaPolicyUpdated {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.enabled {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.EventMFAPolicyUpdated", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.enabled {
            struct_ser.serialize_field("enabled", &self.enabled)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventMfaPolicyUpdated {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "enabled",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            Enabled,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "enabled" => Ok(GeneratedField::Enabled),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventMfaPolicyUpdated;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.EventMFAPolicyUpdated")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventMfaPolicyUpdated, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut enabled__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Enabled => {
                            if enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enabled"));
                            }
                            enabled__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventMfaPolicyUpdated {
                    account_address: account_address__.unwrap_or_default(),
                    enabled: enabled__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.EventMFAPolicyUpdated", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Fido2ChallengeData {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge.is_empty() {
            len += 1;
        }
        if !self.relying_party_id.is_empty() {
            len += 1;
        }
        if !self.allowed_credentials.is_empty() {
            len += 1;
        }
        if !self.user_verification.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.FIDO2ChallengeData", len)?;
        if !self.challenge.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challenge", pbjson::private::base64::encode(&self.challenge).as_str())?;
        }
        if !self.relying_party_id.is_empty() {
            struct_ser.serialize_field("relyingPartyId", &self.relying_party_id)?;
        }
        if !self.allowed_credentials.is_empty() {
            struct_ser.serialize_field("allowedCredentials", &self.allowed_credentials.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.user_verification.is_empty() {
            struct_ser.serialize_field("userVerification", &self.user_verification)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Fido2ChallengeData {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge",
            "relying_party_id",
            "relyingPartyId",
            "allowed_credentials",
            "allowedCredentials",
            "user_verification",
            "userVerification",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Challenge,
            RelyingPartyId,
            AllowedCredentials,
            UserVerification,
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
                            "challenge" => Ok(GeneratedField::Challenge),
                            "relyingPartyId" | "relying_party_id" => Ok(GeneratedField::RelyingPartyId),
                            "allowedCredentials" | "allowed_credentials" => Ok(GeneratedField::AllowedCredentials),
                            "userVerification" | "user_verification" => Ok(GeneratedField::UserVerification),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Fido2ChallengeData;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.FIDO2ChallengeData")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Fido2ChallengeData, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge__ = None;
                let mut relying_party_id__ = None;
                let mut allowed_credentials__ = None;
                let mut user_verification__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Challenge => {
                            if challenge__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challenge"));
                            }
                            challenge__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RelyingPartyId => {
                            if relying_party_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("relyingPartyId"));
                            }
                            relying_party_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowedCredentials => {
                            if allowed_credentials__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowedCredentials"));
                            }
                            allowed_credentials__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::UserVerification => {
                            if user_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userVerification"));
                            }
                            user_verification__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Fido2ChallengeData {
                    challenge: challenge__.unwrap_or_default(),
                    relying_party_id: relying_party_id__.unwrap_or_default(),
                    allowed_credentials: allowed_credentials__.unwrap_or_default(),
                    user_verification: user_verification__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.FIDO2ChallengeData", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Fido2CredentialInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.credential_id.is_empty() {
            len += 1;
        }
        if !self.public_key.is_empty() {
            len += 1;
        }
        if !self.aaguid.is_empty() {
            len += 1;
        }
        if self.sign_count != 0 {
            len += 1;
        }
        if !self.attestation_type.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.FIDO2CredentialInfo", len)?;
        if !self.credential_id.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("credentialId", pbjson::private::base64::encode(&self.credential_id).as_str())?;
        }
        if !self.public_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("publicKey", pbjson::private::base64::encode(&self.public_key).as_str())?;
        }
        if !self.aaguid.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("aaguid", pbjson::private::base64::encode(&self.aaguid).as_str())?;
        }
        if self.sign_count != 0 {
            struct_ser.serialize_field("signCount", &self.sign_count)?;
        }
        if !self.attestation_type.is_empty() {
            struct_ser.serialize_field("attestationType", &self.attestation_type)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Fido2CredentialInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "credential_id",
            "credentialId",
            "public_key",
            "publicKey",
            "aaguid",
            "sign_count",
            "signCount",
            "attestation_type",
            "attestationType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CredentialId,
            PublicKey,
            Aaguid,
            SignCount,
            AttestationType,
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
                            "credentialId" | "credential_id" => Ok(GeneratedField::CredentialId),
                            "publicKey" | "public_key" => Ok(GeneratedField::PublicKey),
                            "aaguid" => Ok(GeneratedField::Aaguid),
                            "signCount" | "sign_count" => Ok(GeneratedField::SignCount),
                            "attestationType" | "attestation_type" => Ok(GeneratedField::AttestationType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Fido2CredentialInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.FIDO2CredentialInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Fido2CredentialInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut credential_id__ = None;
                let mut public_key__ = None;
                let mut aaguid__ = None;
                let mut sign_count__ = None;
                let mut attestation_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CredentialId => {
                            if credential_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("credentialId"));
                            }
                            credential_id__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PublicKey => {
                            if public_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicKey"));
                            }
                            public_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Aaguid => {
                            if aaguid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("aaguid"));
                            }
                            aaguid__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SignCount => {
                            if sign_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signCount"));
                            }
                            sign_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationType => {
                            if attestation_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationType"));
                            }
                            attestation_type__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Fido2CredentialInfo {
                    credential_id: credential_id__.unwrap_or_default(),
                    public_key: public_key__.unwrap_or_default(),
                    aaguid: aaguid__.unwrap_or_default(),
                    sign_count: sign_count__.unwrap_or_default(),
                    attestation_type: attestation_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.FIDO2CredentialInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FactorCombination {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.factors.is_empty() {
            len += 1;
        }
        if self.min_security_level != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.FactorCombination", len)?;
        if !self.factors.is_empty() {
            let v = self.factors.iter().cloned().map(|v| {
                FactorType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("factors", &v)?;
        }
        if self.min_security_level != 0 {
            let v = FactorSecurityLevel::try_from(self.min_security_level)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.min_security_level)))?;
            struct_ser.serialize_field("minSecurityLevel", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for FactorCombination {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "factors",
            "min_security_level",
            "minSecurityLevel",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Factors,
            MinSecurityLevel,
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
                            "factors" => Ok(GeneratedField::Factors),
                            "minSecurityLevel" | "min_security_level" => Ok(GeneratedField::MinSecurityLevel),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FactorCombination;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.FactorCombination")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<FactorCombination, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut factors__ = None;
                let mut min_security_level__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Factors => {
                            if factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factors"));
                            }
                            factors__ = Some(map_.next_value::<Vec<FactorType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                        GeneratedField::MinSecurityLevel => {
                            if min_security_level__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minSecurityLevel"));
                            }
                            min_security_level__ = Some(map_.next_value::<FactorSecurityLevel>()? as i32);
                        }
                    }
                }
                Ok(FactorCombination {
                    factors: factors__.unwrap_or_default(),
                    min_security_level: min_security_level__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.FactorCombination", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FactorEnrollment {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.factor_type != 0 {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        if !self.public_identifier.is_empty() {
            len += 1;
        }
        if !self.label.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.enrolled_at != 0 {
            len += 1;
        }
        if self.verified_at != 0 {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        if self.last_used_at != 0 {
            len += 1;
        }
        if self.use_count != 0 {
            len += 1;
        }
        if self.metadata.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.FactorEnrollment", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.factor_type != 0 {
            let v = FactorType::try_from(self.factor_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type)))?;
            struct_ser.serialize_field("factorType", &v)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        if !self.public_identifier.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("publicIdentifier", pbjson::private::base64::encode(&self.public_identifier).as_str())?;
        }
        if !self.label.is_empty() {
            struct_ser.serialize_field("label", &self.label)?;
        }
        if self.status != 0 {
            let v = FactorEnrollmentStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.enrolled_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("enrolledAt", ToString::to_string(&self.enrolled_at).as_str())?;
        }
        if self.verified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("verifiedAt", ToString::to_string(&self.verified_at).as_str())?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        if self.last_used_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUsedAt", ToString::to_string(&self.last_used_at).as_str())?;
        }
        if self.use_count != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("useCount", ToString::to_string(&self.use_count).as_str())?;
        }
        if let Some(v) = self.metadata.as_ref() {
            struct_ser.serialize_field("metadata", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for FactorEnrollment {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "factor_type",
            "factorType",
            "factor_id",
            "factorId",
            "public_identifier",
            "publicIdentifier",
            "label",
            "status",
            "enrolled_at",
            "enrolledAt",
            "verified_at",
            "verifiedAt",
            "revoked_at",
            "revokedAt",
            "last_used_at",
            "lastUsedAt",
            "use_count",
            "useCount",
            "metadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            FactorType,
            FactorId,
            PublicIdentifier,
            Label,
            Status,
            EnrolledAt,
            VerifiedAt,
            RevokedAt,
            LastUsedAt,
            UseCount,
            Metadata,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            "publicIdentifier" | "public_identifier" => Ok(GeneratedField::PublicIdentifier),
                            "label" => Ok(GeneratedField::Label),
                            "status" => Ok(GeneratedField::Status),
                            "enrolledAt" | "enrolled_at" => Ok(GeneratedField::EnrolledAt),
                            "verifiedAt" | "verified_at" => Ok(GeneratedField::VerifiedAt),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            "lastUsedAt" | "last_used_at" => Ok(GeneratedField::LastUsedAt),
                            "useCount" | "use_count" => Ok(GeneratedField::UseCount),
                            "metadata" => Ok(GeneratedField::Metadata),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FactorEnrollment;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.FactorEnrollment")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<FactorEnrollment, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut factor_type__ = None;
                let mut factor_id__ = None;
                let mut public_identifier__ = None;
                let mut label__ = None;
                let mut status__ = None;
                let mut enrolled_at__ = None;
                let mut verified_at__ = None;
                let mut revoked_at__ = None;
                let mut last_used_at__ = None;
                let mut use_count__ = None;
                let mut metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicIdentifier => {
                            if public_identifier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicIdentifier"));
                            }
                            public_identifier__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Label => {
                            if label__.is_some() {
                                return Err(serde::de::Error::duplicate_field("label"));
                            }
                            label__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<FactorEnrollmentStatus>()? as i32);
                        }
                        GeneratedField::EnrolledAt => {
                            if enrolled_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enrolledAt"));
                            }
                            enrolled_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VerifiedAt => {
                            if verified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedAt"));
                            }
                            verified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastUsedAt => {
                            if last_used_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUsedAt"));
                            }
                            last_used_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UseCount => {
                            if use_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("useCount"));
                            }
                            use_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = map_.next_value()?;
                        }
                    }
                }
                Ok(FactorEnrollment {
                    account_address: account_address__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                    public_identifier: public_identifier__.unwrap_or_default(),
                    label: label__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    enrolled_at: enrolled_at__.unwrap_or_default(),
                    verified_at: verified_at__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                    last_used_at: last_used_at__.unwrap_or_default(),
                    use_count: use_count__.unwrap_or_default(),
                    metadata: metadata__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.FactorEnrollment", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FactorEnrollmentStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::EnrollmentStatusUnspecified => "ENROLLMENT_STATUS_UNSPECIFIED",
            Self::EnrollmentStatusPending => "ENROLLMENT_STATUS_PENDING",
            Self::EnrollmentStatusActive => "ENROLLMENT_STATUS_ACTIVE",
            Self::EnrollmentStatusRevoked => "ENROLLMENT_STATUS_REVOKED",
            Self::EnrollmentStatusExpired => "ENROLLMENT_STATUS_EXPIRED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for FactorEnrollmentStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ENROLLMENT_STATUS_UNSPECIFIED",
            "ENROLLMENT_STATUS_PENDING",
            "ENROLLMENT_STATUS_ACTIVE",
            "ENROLLMENT_STATUS_REVOKED",
            "ENROLLMENT_STATUS_EXPIRED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FactorEnrollmentStatus;

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
                    "ENROLLMENT_STATUS_UNSPECIFIED" => Ok(FactorEnrollmentStatus::EnrollmentStatusUnspecified),
                    "ENROLLMENT_STATUS_PENDING" => Ok(FactorEnrollmentStatus::EnrollmentStatusPending),
                    "ENROLLMENT_STATUS_ACTIVE" => Ok(FactorEnrollmentStatus::EnrollmentStatusActive),
                    "ENROLLMENT_STATUS_REVOKED" => Ok(FactorEnrollmentStatus::EnrollmentStatusRevoked),
                    "ENROLLMENT_STATUS_EXPIRED" => Ok(FactorEnrollmentStatus::EnrollmentStatusExpired),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for FactorMetadata {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.veid_threshold != 0 {
            len += 1;
        }
        if self.device_info.is_some() {
            len += 1;
        }
        if self.fido2_info.is_some() {
            len += 1;
        }
        if !self.contact_hash.is_empty() {
            len += 1;
        }
        if self.hardware_key_info.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.FactorMetadata", len)?;
        if self.veid_threshold != 0 {
            struct_ser.serialize_field("veidThreshold", &self.veid_threshold)?;
        }
        if let Some(v) = self.device_info.as_ref() {
            struct_ser.serialize_field("deviceInfo", v)?;
        }
        if let Some(v) = self.fido2_info.as_ref() {
            struct_ser.serialize_field("fido2Info", v)?;
        }
        if !self.contact_hash.is_empty() {
            struct_ser.serialize_field("contactHash", &self.contact_hash)?;
        }
        if let Some(v) = self.hardware_key_info.as_ref() {
            struct_ser.serialize_field("hardwareKeyInfo", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for FactorMetadata {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "veid_threshold",
            "veidThreshold",
            "device_info",
            "deviceInfo",
            "fido2_info",
            "fido2Info",
            "contact_hash",
            "contactHash",
            "hardware_key_info",
            "hardwareKeyInfo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            VeidThreshold,
            DeviceInfo,
            Fido2Info,
            ContactHash,
            HardwareKeyInfo,
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
                            "veidThreshold" | "veid_threshold" => Ok(GeneratedField::VeidThreshold),
                            "deviceInfo" | "device_info" => Ok(GeneratedField::DeviceInfo),
                            "fido2Info" | "fido2_info" => Ok(GeneratedField::Fido2Info),
                            "contactHash" | "contact_hash" => Ok(GeneratedField::ContactHash),
                            "hardwareKeyInfo" | "hardware_key_info" => Ok(GeneratedField::HardwareKeyInfo),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FactorMetadata;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.FactorMetadata")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<FactorMetadata, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut veid_threshold__ = None;
                let mut device_info__ = None;
                let mut fido2_info__ = None;
                let mut contact_hash__ = None;
                let mut hardware_key_info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::VeidThreshold => {
                            if veid_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidThreshold"));
                            }
                            veid_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DeviceInfo => {
                            if device_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceInfo"));
                            }
                            device_info__ = map_.next_value()?;
                        }
                        GeneratedField::Fido2Info => {
                            if fido2_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fido2Info"));
                            }
                            fido2_info__ = map_.next_value()?;
                        }
                        GeneratedField::ContactHash => {
                            if contact_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("contactHash"));
                            }
                            contact_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HardwareKeyInfo => {
                            if hardware_key_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hardwareKeyInfo"));
                            }
                            hardware_key_info__ = map_.next_value()?;
                        }
                    }
                }
                Ok(FactorMetadata {
                    veid_threshold: veid_threshold__.unwrap_or_default(),
                    device_info: device_info__,
                    fido2_info: fido2_info__,
                    contact_hash: contact_hash__.unwrap_or_default(),
                    hardware_key_info: hardware_key_info__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.FactorMetadata", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for FactorSecurityLevel {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "FACTOR_SECURITY_LEVEL_UNSPECIFIED",
            Self::Low => "FACTOR_SECURITY_LEVEL_LOW",
            Self::Medium => "FACTOR_SECURITY_LEVEL_MEDIUM",
            Self::High => "FACTOR_SECURITY_LEVEL_HIGH",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for FactorSecurityLevel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "FACTOR_SECURITY_LEVEL_UNSPECIFIED",
            "FACTOR_SECURITY_LEVEL_LOW",
            "FACTOR_SECURITY_LEVEL_MEDIUM",
            "FACTOR_SECURITY_LEVEL_HIGH",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FactorSecurityLevel;

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
                    "FACTOR_SECURITY_LEVEL_UNSPECIFIED" => Ok(FactorSecurityLevel::Unspecified),
                    "FACTOR_SECURITY_LEVEL_LOW" => Ok(FactorSecurityLevel::Low),
                    "FACTOR_SECURITY_LEVEL_MEDIUM" => Ok(FactorSecurityLevel::Medium),
                    "FACTOR_SECURITY_LEVEL_HIGH" => Ok(FactorSecurityLevel::High),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for FactorType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "FACTOR_TYPE_UNSPECIFIED",
            Self::Totp => "FACTOR_TYPE_TOTP",
            Self::Fido2 => "FACTOR_TYPE_FIDO2",
            Self::Sms => "FACTOR_TYPE_SMS",
            Self::Email => "FACTOR_TYPE_EMAIL",
            Self::Veid => "FACTOR_TYPE_VEID",
            Self::TrustedDevice => "FACTOR_TYPE_TRUSTED_DEVICE",
            Self::HardwareKey => "FACTOR_TYPE_HARDWARE_KEY",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for FactorType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "FACTOR_TYPE_UNSPECIFIED",
            "FACTOR_TYPE_TOTP",
            "FACTOR_TYPE_FIDO2",
            "FACTOR_TYPE_SMS",
            "FACTOR_TYPE_EMAIL",
            "FACTOR_TYPE_VEID",
            "FACTOR_TYPE_TRUSTED_DEVICE",
            "FACTOR_TYPE_HARDWARE_KEY",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = FactorType;

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
                    "FACTOR_TYPE_UNSPECIFIED" => Ok(FactorType::Unspecified),
                    "FACTOR_TYPE_TOTP" => Ok(FactorType::Totp),
                    "FACTOR_TYPE_FIDO2" => Ok(FactorType::Fido2),
                    "FACTOR_TYPE_SMS" => Ok(FactorType::Sms),
                    "FACTOR_TYPE_EMAIL" => Ok(FactorType::Email),
                    "FACTOR_TYPE_VEID" => Ok(FactorType::Veid),
                    "FACTOR_TYPE_TRUSTED_DEVICE" => Ok(FactorType::TrustedDevice),
                    "FACTOR_TYPE_HARDWARE_KEY" => Ok(FactorType::HardwareKey),
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
        if !self.mfa_policies.is_empty() {
            len += 1;
        }
        if !self.factor_enrollments.is_empty() {
            len += 1;
        }
        if !self.sensitive_tx_configs.is_empty() {
            len += 1;
        }
        if !self.trusted_devices.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.mfa_policies.is_empty() {
            struct_ser.serialize_field("mfaPolicies", &self.mfa_policies)?;
        }
        if !self.factor_enrollments.is_empty() {
            struct_ser.serialize_field("factorEnrollments", &self.factor_enrollments)?;
        }
        if !self.sensitive_tx_configs.is_empty() {
            struct_ser.serialize_field("sensitiveTxConfigs", &self.sensitive_tx_configs)?;
        }
        if !self.trusted_devices.is_empty() {
            struct_ser.serialize_field("trustedDevices", &self.trusted_devices)?;
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
            "mfa_policies",
            "mfaPolicies",
            "factor_enrollments",
            "factorEnrollments",
            "sensitive_tx_configs",
            "sensitiveTxConfigs",
            "trusted_devices",
            "trustedDevices",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            MfaPolicies,
            FactorEnrollments,
            SensitiveTxConfigs,
            TrustedDevices,
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
                            "mfaPolicies" | "mfa_policies" => Ok(GeneratedField::MfaPolicies),
                            "factorEnrollments" | "factor_enrollments" => Ok(GeneratedField::FactorEnrollments),
                            "sensitiveTxConfigs" | "sensitive_tx_configs" => Ok(GeneratedField::SensitiveTxConfigs),
                            "trustedDevices" | "trusted_devices" => Ok(GeneratedField::TrustedDevices),
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
                formatter.write_str("struct virtengine.mfa.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut mfa_policies__ = None;
                let mut factor_enrollments__ = None;
                let mut sensitive_tx_configs__ = None;
                let mut trusted_devices__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::MfaPolicies => {
                            if mfa_policies__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaPolicies"));
                            }
                            mfa_policies__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorEnrollments => {
                            if factor_enrollments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorEnrollments"));
                            }
                            factor_enrollments__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SensitiveTxConfigs => {
                            if sensitive_tx_configs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sensitiveTxConfigs"));
                            }
                            sensitive_tx_configs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TrustedDevices => {
                            if trusted_devices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustedDevices"));
                            }
                            trusted_devices__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    mfa_policies: mfa_policies__.unwrap_or_default(),
                    factor_enrollments: factor_enrollments__.unwrap_or_default(),
                    sensitive_tx_configs: sensitive_tx_configs__.unwrap_or_default(),
                    trusted_devices: trusted_devices__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HardwareKeyChallenge {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge.is_empty() {
            len += 1;
        }
        if !self.key_id.is_empty() {
            len += 1;
        }
        if !self.nonce.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.HardwareKeyChallenge", len)?;
        if !self.challenge.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challenge", pbjson::private::base64::encode(&self.challenge).as_str())?;
        }
        if !self.key_id.is_empty() {
            struct_ser.serialize_field("keyId", &self.key_id)?;
        }
        if !self.nonce.is_empty() {
            struct_ser.serialize_field("nonce", &self.nonce)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HardwareKeyChallenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge",
            "key_id",
            "keyId",
            "nonce",
            "created_at",
            "createdAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Challenge,
            KeyId,
            Nonce,
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
                            "challenge" => Ok(GeneratedField::Challenge),
                            "keyId" | "key_id" => Ok(GeneratedField::KeyId),
                            "nonce" => Ok(GeneratedField::Nonce),
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
            type Value = HardwareKeyChallenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.HardwareKeyChallenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HardwareKeyChallenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge__ = None;
                let mut key_id__ = None;
                let mut nonce__ = None;
                let mut created_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Challenge => {
                            if challenge__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challenge"));
                            }
                            challenge__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::KeyId => {
                            if key_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyId"));
                            }
                            key_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Nonce => {
                            if nonce__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nonce"));
                            }
                            nonce__ = Some(map_.next_value()?);
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
                Ok(HardwareKeyChallenge {
                    challenge: challenge__.unwrap_or_default(),
                    key_id: key_id__.unwrap_or_default(),
                    nonce: nonce__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.HardwareKeyChallenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HardwareKeyEnrollment {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.key_type != 0 {
            len += 1;
        }
        if !self.key_id.is_empty() {
            len += 1;
        }
        if !self.subject_dn.is_empty() {
            len += 1;
        }
        if !self.issuer_dn.is_empty() {
            len += 1;
        }
        if !self.serial_number.is_empty() {
            len += 1;
        }
        if !self.public_key_fingerprint.is_empty() {
            len += 1;
        }
        if self.not_before != 0 {
            len += 1;
        }
        if self.not_after != 0 {
            len += 1;
        }
        if !self.key_usage.is_empty() {
            len += 1;
        }
        if !self.extended_key_usage.is_empty() {
            len += 1;
        }
        if self.smart_card_info.is_some() {
            len += 1;
        }
        if self.revocation_check_enabled {
            len += 1;
        }
        if self.last_revocation_check != 0 {
            len += 1;
        }
        if self.revocation_status != 0 {
            len += 1;
        }
        if !self.trusted_ca_cert_fingerprints.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.HardwareKeyEnrollment", len)?;
        if self.key_type != 0 {
            let v = HardwareKeyType::try_from(self.key_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.key_type)))?;
            struct_ser.serialize_field("keyType", &v)?;
        }
        if !self.key_id.is_empty() {
            struct_ser.serialize_field("keyId", &self.key_id)?;
        }
        if !self.subject_dn.is_empty() {
            struct_ser.serialize_field("subjectDn", &self.subject_dn)?;
        }
        if !self.issuer_dn.is_empty() {
            struct_ser.serialize_field("issuerDn", &self.issuer_dn)?;
        }
        if !self.serial_number.is_empty() {
            struct_ser.serialize_field("serialNumber", &self.serial_number)?;
        }
        if !self.public_key_fingerprint.is_empty() {
            struct_ser.serialize_field("publicKeyFingerprint", &self.public_key_fingerprint)?;
        }
        if self.not_before != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("notBefore", ToString::to_string(&self.not_before).as_str())?;
        }
        if self.not_after != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("notAfter", ToString::to_string(&self.not_after).as_str())?;
        }
        if !self.key_usage.is_empty() {
            struct_ser.serialize_field("keyUsage", &self.key_usage)?;
        }
        if !self.extended_key_usage.is_empty() {
            struct_ser.serialize_field("extendedKeyUsage", &self.extended_key_usage)?;
        }
        if let Some(v) = self.smart_card_info.as_ref() {
            struct_ser.serialize_field("smartCardInfo", v)?;
        }
        if self.revocation_check_enabled {
            struct_ser.serialize_field("revocationCheckEnabled", &self.revocation_check_enabled)?;
        }
        if self.last_revocation_check != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastRevocationCheck", ToString::to_string(&self.last_revocation_check).as_str())?;
        }
        if self.revocation_status != 0 {
            let v = RevocationStatus::try_from(self.revocation_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.revocation_status)))?;
            struct_ser.serialize_field("revocationStatus", &v)?;
        }
        if !self.trusted_ca_cert_fingerprints.is_empty() {
            struct_ser.serialize_field("trustedCaCertFingerprints", &self.trusted_ca_cert_fingerprints)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HardwareKeyEnrollment {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key_type",
            "keyType",
            "key_id",
            "keyId",
            "subject_dn",
            "subjectDn",
            "issuer_dn",
            "issuerDn",
            "serial_number",
            "serialNumber",
            "public_key_fingerprint",
            "publicKeyFingerprint",
            "not_before",
            "notBefore",
            "not_after",
            "notAfter",
            "key_usage",
            "keyUsage",
            "extended_key_usage",
            "extendedKeyUsage",
            "smart_card_info",
            "smartCardInfo",
            "revocation_check_enabled",
            "revocationCheckEnabled",
            "last_revocation_check",
            "lastRevocationCheck",
            "revocation_status",
            "revocationStatus",
            "trusted_ca_cert_fingerprints",
            "trustedCaCertFingerprints",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            KeyType,
            KeyId,
            SubjectDn,
            IssuerDn,
            SerialNumber,
            PublicKeyFingerprint,
            NotBefore,
            NotAfter,
            KeyUsage,
            ExtendedKeyUsage,
            SmartCardInfo,
            RevocationCheckEnabled,
            LastRevocationCheck,
            RevocationStatus,
            TrustedCaCertFingerprints,
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
                            "keyType" | "key_type" => Ok(GeneratedField::KeyType),
                            "keyId" | "key_id" => Ok(GeneratedField::KeyId),
                            "subjectDn" | "subject_dn" => Ok(GeneratedField::SubjectDn),
                            "issuerDn" | "issuer_dn" => Ok(GeneratedField::IssuerDn),
                            "serialNumber" | "serial_number" => Ok(GeneratedField::SerialNumber),
                            "publicKeyFingerprint" | "public_key_fingerprint" => Ok(GeneratedField::PublicKeyFingerprint),
                            "notBefore" | "not_before" => Ok(GeneratedField::NotBefore),
                            "notAfter" | "not_after" => Ok(GeneratedField::NotAfter),
                            "keyUsage" | "key_usage" => Ok(GeneratedField::KeyUsage),
                            "extendedKeyUsage" | "extended_key_usage" => Ok(GeneratedField::ExtendedKeyUsage),
                            "smartCardInfo" | "smart_card_info" => Ok(GeneratedField::SmartCardInfo),
                            "revocationCheckEnabled" | "revocation_check_enabled" => Ok(GeneratedField::RevocationCheckEnabled),
                            "lastRevocationCheck" | "last_revocation_check" => Ok(GeneratedField::LastRevocationCheck),
                            "revocationStatus" | "revocation_status" => Ok(GeneratedField::RevocationStatus),
                            "trustedCaCertFingerprints" | "trusted_ca_cert_fingerprints" => Ok(GeneratedField::TrustedCaCertFingerprints),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HardwareKeyEnrollment;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.HardwareKeyEnrollment")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HardwareKeyEnrollment, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key_type__ = None;
                let mut key_id__ = None;
                let mut subject_dn__ = None;
                let mut issuer_dn__ = None;
                let mut serial_number__ = None;
                let mut public_key_fingerprint__ = None;
                let mut not_before__ = None;
                let mut not_after__ = None;
                let mut key_usage__ = None;
                let mut extended_key_usage__ = None;
                let mut smart_card_info__ = None;
                let mut revocation_check_enabled__ = None;
                let mut last_revocation_check__ = None;
                let mut revocation_status__ = None;
                let mut trusted_ca_cert_fingerprints__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::KeyType => {
                            if key_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyType"));
                            }
                            key_type__ = Some(map_.next_value::<HardwareKeyType>()? as i32);
                        }
                        GeneratedField::KeyId => {
                            if key_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyId"));
                            }
                            key_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubjectDn => {
                            if subject_dn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("subjectDn"));
                            }
                            subject_dn__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IssuerDn => {
                            if issuer_dn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("issuerDn"));
                            }
                            issuer_dn__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SerialNumber => {
                            if serial_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("serialNumber"));
                            }
                            serial_number__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicKeyFingerprint => {
                            if public_key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicKeyFingerprint"));
                            }
                            public_key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NotBefore => {
                            if not_before__.is_some() {
                                return Err(serde::de::Error::duplicate_field("notBefore"));
                            }
                            not_before__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NotAfter => {
                            if not_after__.is_some() {
                                return Err(serde::de::Error::duplicate_field("notAfter"));
                            }
                            not_after__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::KeyUsage => {
                            if key_usage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyUsage"));
                            }
                            key_usage__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExtendedKeyUsage => {
                            if extended_key_usage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("extendedKeyUsage"));
                            }
                            extended_key_usage__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SmartCardInfo => {
                            if smart_card_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("smartCardInfo"));
                            }
                            smart_card_info__ = map_.next_value()?;
                        }
                        GeneratedField::RevocationCheckEnabled => {
                            if revocation_check_enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revocationCheckEnabled"));
                            }
                            revocation_check_enabled__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastRevocationCheck => {
                            if last_revocation_check__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastRevocationCheck"));
                            }
                            last_revocation_check__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RevocationStatus => {
                            if revocation_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revocationStatus"));
                            }
                            revocation_status__ = Some(map_.next_value::<RevocationStatus>()? as i32);
                        }
                        GeneratedField::TrustedCaCertFingerprints => {
                            if trusted_ca_cert_fingerprints__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustedCaCertFingerprints"));
                            }
                            trusted_ca_cert_fingerprints__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(HardwareKeyEnrollment {
                    key_type: key_type__.unwrap_or_default(),
                    key_id: key_id__.unwrap_or_default(),
                    subject_dn: subject_dn__.unwrap_or_default(),
                    issuer_dn: issuer_dn__.unwrap_or_default(),
                    serial_number: serial_number__.unwrap_or_default(),
                    public_key_fingerprint: public_key_fingerprint__.unwrap_or_default(),
                    not_before: not_before__.unwrap_or_default(),
                    not_after: not_after__.unwrap_or_default(),
                    key_usage: key_usage__.unwrap_or_default(),
                    extended_key_usage: extended_key_usage__.unwrap_or_default(),
                    smart_card_info: smart_card_info__,
                    revocation_check_enabled: revocation_check_enabled__.unwrap_or_default(),
                    last_revocation_check: last_revocation_check__.unwrap_or_default(),
                    revocation_status: revocation_status__.unwrap_or_default(),
                    trusted_ca_cert_fingerprints: trusted_ca_cert_fingerprints__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.HardwareKeyEnrollment", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HardwareKeyType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "HARDWARE_KEY_TYPE_UNSPECIFIED",
            Self::X509 => "HARDWARE_KEY_TYPE_X509",
            Self::SmartCard => "HARDWARE_KEY_TYPE_SMART_CARD",
            Self::Piv => "HARDWARE_KEY_TYPE_PIV",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for HardwareKeyType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "HARDWARE_KEY_TYPE_UNSPECIFIED",
            "HARDWARE_KEY_TYPE_X509",
            "HARDWARE_KEY_TYPE_SMART_CARD",
            "HARDWARE_KEY_TYPE_PIV",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HardwareKeyType;

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
                    "HARDWARE_KEY_TYPE_UNSPECIFIED" => Ok(HardwareKeyType::Unspecified),
                    "HARDWARE_KEY_TYPE_X509" => Ok(HardwareKeyType::X509),
                    "HARDWARE_KEY_TYPE_SMART_CARD" => Ok(HardwareKeyType::SmartCard),
                    "HARDWARE_KEY_TYPE_PIV" => Ok(HardwareKeyType::Piv),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for MfaPolicy {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.required_factors.is_empty() {
            len += 1;
        }
        if self.trusted_device_rule.is_some() {
            len += 1;
        }
        if !self.recovery_factors.is_empty() {
            len += 1;
        }
        if !self.key_rotation_factors.is_empty() {
            len += 1;
        }
        if self.session_duration != 0 {
            len += 1;
        }
        if self.veid_threshold != 0 {
            len += 1;
        }
        if self.enabled {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MFAPolicy", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.required_factors.is_empty() {
            struct_ser.serialize_field("requiredFactors", &self.required_factors)?;
        }
        if let Some(v) = self.trusted_device_rule.as_ref() {
            struct_ser.serialize_field("trustedDeviceRule", v)?;
        }
        if !self.recovery_factors.is_empty() {
            struct_ser.serialize_field("recoveryFactors", &self.recovery_factors)?;
        }
        if !self.key_rotation_factors.is_empty() {
            struct_ser.serialize_field("keyRotationFactors", &self.key_rotation_factors)?;
        }
        if self.session_duration != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sessionDuration", ToString::to_string(&self.session_duration).as_str())?;
        }
        if self.veid_threshold != 0 {
            struct_ser.serialize_field("veidThreshold", &self.veid_threshold)?;
        }
        if self.enabled {
            struct_ser.serialize_field("enabled", &self.enabled)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MfaPolicy {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "required_factors",
            "requiredFactors",
            "trusted_device_rule",
            "trustedDeviceRule",
            "recovery_factors",
            "recoveryFactors",
            "key_rotation_factors",
            "keyRotationFactors",
            "session_duration",
            "sessionDuration",
            "veid_threshold",
            "veidThreshold",
            "enabled",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            RequiredFactors,
            TrustedDeviceRule,
            RecoveryFactors,
            KeyRotationFactors,
            SessionDuration,
            VeidThreshold,
            Enabled,
            CreatedAt,
            UpdatedAt,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "requiredFactors" | "required_factors" => Ok(GeneratedField::RequiredFactors),
                            "trustedDeviceRule" | "trusted_device_rule" => Ok(GeneratedField::TrustedDeviceRule),
                            "recoveryFactors" | "recovery_factors" => Ok(GeneratedField::RecoveryFactors),
                            "keyRotationFactors" | "key_rotation_factors" => Ok(GeneratedField::KeyRotationFactors),
                            "sessionDuration" | "session_duration" => Ok(GeneratedField::SessionDuration),
                            "veidThreshold" | "veid_threshold" => Ok(GeneratedField::VeidThreshold),
                            "enabled" => Ok(GeneratedField::Enabled),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MfaPolicy;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MFAPolicy")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MfaPolicy, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut required_factors__ = None;
                let mut trusted_device_rule__ = None;
                let mut recovery_factors__ = None;
                let mut key_rotation_factors__ = None;
                let mut session_duration__ = None;
                let mut veid_threshold__ = None;
                let mut enabled__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequiredFactors => {
                            if required_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredFactors"));
                            }
                            required_factors__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TrustedDeviceRule => {
                            if trusted_device_rule__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustedDeviceRule"));
                            }
                            trusted_device_rule__ = map_.next_value()?;
                        }
                        GeneratedField::RecoveryFactors => {
                            if recovery_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recoveryFactors"));
                            }
                            recovery_factors__ = Some(map_.next_value()?);
                        }
                        GeneratedField::KeyRotationFactors => {
                            if key_rotation_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyRotationFactors"));
                            }
                            key_rotation_factors__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SessionDuration => {
                            if session_duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionDuration"));
                            }
                            session_duration__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidThreshold => {
                            if veid_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidThreshold"));
                            }
                            veid_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Enabled => {
                            if enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enabled"));
                            }
                            enabled__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MfaPolicy {
                    account_address: account_address__.unwrap_or_default(),
                    required_factors: required_factors__.unwrap_or_default(),
                    trusted_device_rule: trusted_device_rule__,
                    recovery_factors: recovery_factors__.unwrap_or_default(),
                    key_rotation_factors: key_rotation_factors__.unwrap_or_default(),
                    session_duration: session_duration__.unwrap_or_default(),
                    veid_threshold: veid_threshold__.unwrap_or_default(),
                    enabled: enabled__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MFAPolicy", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MfaProof {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.session_id.is_empty() {
            len += 1;
        }
        if !self.verified_factors.is_empty() {
            len += 1;
        }
        if self.timestamp != 0 {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MFAProof", len)?;
        if !self.session_id.is_empty() {
            struct_ser.serialize_field("sessionId", &self.session_id)?;
        }
        if !self.verified_factors.is_empty() {
            let v = self.verified_factors.iter().cloned().map(|v| {
                FactorType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("verifiedFactors", &v)?;
        }
        if self.timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("timestamp", ToString::to_string(&self.timestamp).as_str())?;
        }
        if !self.signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signature", pbjson::private::base64::encode(&self.signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MfaProof {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "session_id",
            "sessionId",
            "verified_factors",
            "verifiedFactors",
            "timestamp",
            "signature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SessionId,
            VerifiedFactors,
            Timestamp,
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
                            "sessionId" | "session_id" => Ok(GeneratedField::SessionId),
                            "verifiedFactors" | "verified_factors" => Ok(GeneratedField::VerifiedFactors),
                            "timestamp" => Ok(GeneratedField::Timestamp),
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
            type Value = MfaProof;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MFAProof")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MfaProof, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut session_id__ = None;
                let mut verified_factors__ = None;
                let mut timestamp__ = None;
                let mut signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SessionId => {
                            if session_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionId"));
                            }
                            session_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::VerifiedFactors => {
                            if verified_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedFactors"));
                            }
                            verified_factors__ = Some(map_.next_value::<Vec<FactorType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
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
                Ok(MfaProof {
                    session_id: session_id__.unwrap_or_default(),
                    verified_factors: verified_factors__.unwrap_or_default(),
                    timestamp: timestamp__.unwrap_or_default(),
                    signature: signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MFAProof", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAddTrustedDevice {
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
        if self.device_info.is_some() {
            len += 1;
        }
        if self.mfa_proof.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgAddTrustedDevice", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if let Some(v) = self.device_info.as_ref() {
            struct_ser.serialize_field("deviceInfo", v)?;
        }
        if let Some(v) = self.mfa_proof.as_ref() {
            struct_ser.serialize_field("mfaProof", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAddTrustedDevice {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "device_info",
            "deviceInfo",
            "mfa_proof",
            "mfaProof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            DeviceInfo,
            MfaProof,
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
                            "deviceInfo" | "device_info" => Ok(GeneratedField::DeviceInfo),
                            "mfaProof" | "mfa_proof" => Ok(GeneratedField::MfaProof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAddTrustedDevice;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgAddTrustedDevice")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAddTrustedDevice, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut device_info__ = None;
                let mut mfa_proof__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DeviceInfo => {
                            if device_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceInfo"));
                            }
                            device_info__ = map_.next_value()?;
                        }
                        GeneratedField::MfaProof => {
                            if mfa_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaProof"));
                            }
                            mfa_proof__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgAddTrustedDevice {
                    sender: sender__.unwrap_or_default(),
                    device_info: device_info__,
                    mfa_proof: mfa_proof__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgAddTrustedDevice", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAddTrustedDeviceResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.success {
            len += 1;
        }
        if self.trust_expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgAddTrustedDeviceResponse", len)?;
        if self.success {
            struct_ser.serialize_field("success", &self.success)?;
        }
        if self.trust_expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("trustExpiresAt", ToString::to_string(&self.trust_expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAddTrustedDeviceResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "success",
            "trust_expires_at",
            "trustExpiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Success,
            TrustExpiresAt,
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
                            "success" => Ok(GeneratedField::Success),
                            "trustExpiresAt" | "trust_expires_at" => Ok(GeneratedField::TrustExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAddTrustedDeviceResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgAddTrustedDeviceResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAddTrustedDeviceResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut success__ = None;
                let mut trust_expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Success => {
                            if success__.is_some() {
                                return Err(serde::de::Error::duplicate_field("success"));
                            }
                            success__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TrustExpiresAt => {
                            if trust_expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustExpiresAt"));
                            }
                            trust_expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAddTrustedDeviceResponse {
                    success: success__.unwrap_or_default(),
                    trust_expires_at: trust_expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgAddTrustedDeviceResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateChallenge {
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
        if self.factor_type != 0 {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        if self.transaction_type != 0 {
            len += 1;
        }
        if self.client_info.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgCreateChallenge", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if self.factor_type != 0 {
            let v = FactorType::try_from(self.factor_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type)))?;
            struct_ser.serialize_field("factorType", &v)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        if self.transaction_type != 0 {
            let v = SensitiveTransactionType::try_from(self.transaction_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.transaction_type)))?;
            struct_ser.serialize_field("transactionType", &v)?;
        }
        if let Some(v) = self.client_info.as_ref() {
            struct_ser.serialize_field("clientInfo", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateChallenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "factor_type",
            "factorType",
            "factor_id",
            "factorId",
            "transaction_type",
            "transactionType",
            "client_info",
            "clientInfo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            FactorType,
            FactorId,
            TransactionType,
            ClientInfo,
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
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            "clientInfo" | "client_info" => Ok(GeneratedField::ClientInfo),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateChallenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgCreateChallenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateChallenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut factor_type__ = None;
                let mut factor_id__ = None;
                let mut transaction_type__ = None;
                let mut client_info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value::<SensitiveTransactionType>()? as i32);
                        }
                        GeneratedField::ClientInfo => {
                            if client_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientInfo"));
                            }
                            client_info__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgCreateChallenge {
                    sender: sender__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                    transaction_type: transaction_type__.unwrap_or_default(),
                    client_info: client_info__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgCreateChallenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateChallengeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.challenge_data.is_empty() {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgCreateChallengeResponse", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.challenge_data.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challengeData", pbjson::private::base64::encode(&self.challenge_data).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateChallengeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "challenge_data",
            "challengeData",
            "expires_at",
            "expiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            ChallengeData,
            ExpiresAt,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "challengeData" | "challenge_data" => Ok(GeneratedField::ChallengeData),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateChallengeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgCreateChallengeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateChallengeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut challenge_data__ = None;
                let mut expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChallengeData => {
                            if challenge_data__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeData"));
                            }
                            challenge_data__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgCreateChallengeResponse {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    challenge_data: challenge_data__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgCreateChallengeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgEnrollFactor {
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
        if self.factor_type != 0 {
            len += 1;
        }
        if !self.label.is_empty() {
            len += 1;
        }
        if !self.public_identifier.is_empty() {
            len += 1;
        }
        if self.metadata.is_some() {
            len += 1;
        }
        if !self.initial_verification_proof.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgEnrollFactor", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if self.factor_type != 0 {
            let v = FactorType::try_from(self.factor_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type)))?;
            struct_ser.serialize_field("factorType", &v)?;
        }
        if !self.label.is_empty() {
            struct_ser.serialize_field("label", &self.label)?;
        }
        if !self.public_identifier.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("publicIdentifier", pbjson::private::base64::encode(&self.public_identifier).as_str())?;
        }
        if let Some(v) = self.metadata.as_ref() {
            struct_ser.serialize_field("metadata", v)?;
        }
        if !self.initial_verification_proof.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("initialVerificationProof", pbjson::private::base64::encode(&self.initial_verification_proof).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgEnrollFactor {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "factor_type",
            "factorType",
            "label",
            "public_identifier",
            "publicIdentifier",
            "metadata",
            "initial_verification_proof",
            "initialVerificationProof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            FactorType,
            Label,
            PublicIdentifier,
            Metadata,
            InitialVerificationProof,
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
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "label" => Ok(GeneratedField::Label),
                            "publicIdentifier" | "public_identifier" => Ok(GeneratedField::PublicIdentifier),
                            "metadata" => Ok(GeneratedField::Metadata),
                            "initialVerificationProof" | "initial_verification_proof" => Ok(GeneratedField::InitialVerificationProof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgEnrollFactor;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgEnrollFactor")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgEnrollFactor, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut factor_type__ = None;
                let mut label__ = None;
                let mut public_identifier__ = None;
                let mut metadata__ = None;
                let mut initial_verification_proof__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::Label => {
                            if label__.is_some() {
                                return Err(serde::de::Error::duplicate_field("label"));
                            }
                            label__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicIdentifier => {
                            if public_identifier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicIdentifier"));
                            }
                            public_identifier__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = map_.next_value()?;
                        }
                        GeneratedField::InitialVerificationProof => {
                            if initial_verification_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("initialVerificationProof"));
                            }
                            initial_verification_proof__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgEnrollFactor {
                    sender: sender__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    label: label__.unwrap_or_default(),
                    public_identifier: public_identifier__.unwrap_or_default(),
                    metadata: metadata__,
                    initial_verification_proof: initial_verification_proof__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgEnrollFactor", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgEnrollFactorResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.factor_id.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgEnrollFactorResponse", len)?;
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        if self.status != 0 {
            let v = FactorEnrollmentStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgEnrollFactorResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "factor_id",
            "factorId",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            FactorId,
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
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
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
            type Value = MsgEnrollFactorResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgEnrollFactorResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgEnrollFactorResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut factor_id__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<FactorEnrollmentStatus>()? as i32);
                        }
                    }
                }
                Ok(MsgEnrollFactorResponse {
                    factor_id: factor_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgEnrollFactorResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRemoveTrustedDevice {
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
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        if self.mfa_proof.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgRemoveTrustedDevice", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        if let Some(v) = self.mfa_proof.as_ref() {
            struct_ser.serialize_field("mfaProof", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRemoveTrustedDevice {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "device_fingerprint",
            "deviceFingerprint",
            "mfa_proof",
            "mfaProof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            DeviceFingerprint,
            MfaProof,
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
                            "deviceFingerprint" | "device_fingerprint" => Ok(GeneratedField::DeviceFingerprint),
                            "mfaProof" | "mfa_proof" => Ok(GeneratedField::MfaProof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRemoveTrustedDevice;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgRemoveTrustedDevice")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRemoveTrustedDevice, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut device_fingerprint__ = None;
                let mut mfa_proof__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MfaProof => {
                            if mfa_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaProof"));
                            }
                            mfa_proof__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgRemoveTrustedDevice {
                    sender: sender__.unwrap_or_default(),
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                    mfa_proof: mfa_proof__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgRemoveTrustedDevice", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRemoveTrustedDeviceResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.success {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgRemoveTrustedDeviceResponse", len)?;
        if self.success {
            struct_ser.serialize_field("success", &self.success)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRemoveTrustedDeviceResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "success",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Success,
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
                            "success" => Ok(GeneratedField::Success),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRemoveTrustedDeviceResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgRemoveTrustedDeviceResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRemoveTrustedDeviceResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut success__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Success => {
                            if success__.is_some() {
                                return Err(serde::de::Error::duplicate_field("success"));
                            }
                            success__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRemoveTrustedDeviceResponse {
                    success: success__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgRemoveTrustedDeviceResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeFactor {
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
        if self.factor_type != 0 {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        if self.mfa_proof.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgRevokeFactor", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if self.factor_type != 0 {
            let v = FactorType::try_from(self.factor_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type)))?;
            struct_ser.serialize_field("factorType", &v)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        if let Some(v) = self.mfa_proof.as_ref() {
            struct_ser.serialize_field("mfaProof", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeFactor {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "factor_type",
            "factorType",
            "factor_id",
            "factorId",
            "mfa_proof",
            "mfaProof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            FactorType,
            FactorId,
            MfaProof,
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
                            "factorType" | "factor_type" => Ok(GeneratedField::FactorType),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            "mfaProof" | "mfa_proof" => Ok(GeneratedField::MfaProof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeFactor;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgRevokeFactor")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeFactor, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut factor_type__ = None;
                let mut factor_id__ = None;
                let mut mfa_proof__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorType => {
                            if factor_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorType"));
                            }
                            factor_type__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MfaProof => {
                            if mfa_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaProof"));
                            }
                            mfa_proof__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgRevokeFactor {
                    sender: sender__.unwrap_or_default(),
                    factor_type: factor_type__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                    mfa_proof: mfa_proof__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgRevokeFactor", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeFactorResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.success {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgRevokeFactorResponse", len)?;
        if self.success {
            struct_ser.serialize_field("success", &self.success)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeFactorResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "success",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Success,
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
                            "success" => Ok(GeneratedField::Success),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeFactorResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgRevokeFactorResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeFactorResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut success__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Success => {
                            if success__.is_some() {
                                return Err(serde::de::Error::duplicate_field("success"));
                            }
                            success__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRevokeFactorResponse {
                    success: success__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgRevokeFactorResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSetMfaPolicy {
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
        if self.policy.is_some() {
            len += 1;
        }
        if self.mfa_proof.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgSetMFAPolicy", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if let Some(v) = self.policy.as_ref() {
            struct_ser.serialize_field("policy", v)?;
        }
        if let Some(v) = self.mfa_proof.as_ref() {
            struct_ser.serialize_field("mfaProof", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSetMfaPolicy {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "policy",
            "mfa_proof",
            "mfaProof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            Policy,
            MfaProof,
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
                            "policy" => Ok(GeneratedField::Policy),
                            "mfaProof" | "mfa_proof" => Ok(GeneratedField::MfaProof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSetMfaPolicy;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgSetMFAPolicy")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSetMfaPolicy, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut policy__ = None;
                let mut mfa_proof__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Policy => {
                            if policy__.is_some() {
                                return Err(serde::de::Error::duplicate_field("policy"));
                            }
                            policy__ = map_.next_value()?;
                        }
                        GeneratedField::MfaProof => {
                            if mfa_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaProof"));
                            }
                            mfa_proof__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgSetMfaPolicy {
                    sender: sender__.unwrap_or_default(),
                    policy: policy__,
                    mfa_proof: mfa_proof__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgSetMFAPolicy", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSetMfaPolicyResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.success {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgSetMFAPolicyResponse", len)?;
        if self.success {
            struct_ser.serialize_field("success", &self.success)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSetMfaPolicyResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "success",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Success,
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
                            "success" => Ok(GeneratedField::Success),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSetMfaPolicyResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgSetMFAPolicyResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSetMfaPolicyResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut success__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Success => {
                            if success__.is_some() {
                                return Err(serde::de::Error::duplicate_field("success"));
                            }
                            success__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSetMfaPolicyResponse {
                    success: success__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgSetMFAPolicyResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.mfa.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.mfa.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateSensitiveTxConfig {
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
        if self.config.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgUpdateSensitiveTxConfig", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.config.as_ref() {
            struct_ser.serialize_field("config", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateSensitiveTxConfig {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "config",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            Config,
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
                            "config" => Ok(GeneratedField::Config),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateSensitiveTxConfig;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgUpdateSensitiveTxConfig")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateSensitiveTxConfig, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut config__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Config => {
                            if config__.is_some() {
                                return Err(serde::de::Error::duplicate_field("config"));
                            }
                            config__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgUpdateSensitiveTxConfig {
                    authority: authority__.unwrap_or_default(),
                    config: config__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgUpdateSensitiveTxConfig", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateSensitiveTxConfigResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.success {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgUpdateSensitiveTxConfigResponse", len)?;
        if self.success {
            struct_ser.serialize_field("success", &self.success)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateSensitiveTxConfigResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "success",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Success,
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
                            "success" => Ok(GeneratedField::Success),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateSensitiveTxConfigResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgUpdateSensitiveTxConfigResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateSensitiveTxConfigResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut success__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Success => {
                            if success__.is_some() {
                                return Err(serde::de::Error::duplicate_field("success"));
                            }
                            success__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateSensitiveTxConfigResponse {
                    success: success__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgUpdateSensitiveTxConfigResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgVerifyChallenge {
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
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if self.response.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgVerifyChallenge", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if let Some(v) = self.response.as_ref() {
            struct_ser.serialize_field("response", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgVerifyChallenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "challenge_id",
            "challengeId",
            "response",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ChallengeId,
            Response,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "response" => Ok(GeneratedField::Response),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgVerifyChallenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgVerifyChallenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgVerifyChallenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut challenge_id__ = None;
                let mut response__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Response => {
                            if response__.is_some() {
                                return Err(serde::de::Error::duplicate_field("response"));
                            }
                            response__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgVerifyChallenge {
                    sender: sender__.unwrap_or_default(),
                    challenge_id: challenge_id__.unwrap_or_default(),
                    response: response__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgVerifyChallenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgVerifyChallengeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.verified {
            len += 1;
        }
        if !self.session_id.is_empty() {
            len += 1;
        }
        if self.session_expires_at != 0 {
            len += 1;
        }
        if !self.remaining_factors.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.MsgVerifyChallengeResponse", len)?;
        if self.verified {
            struct_ser.serialize_field("verified", &self.verified)?;
        }
        if !self.session_id.is_empty() {
            struct_ser.serialize_field("sessionId", &self.session_id)?;
        }
        if self.session_expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sessionExpiresAt", ToString::to_string(&self.session_expires_at).as_str())?;
        }
        if !self.remaining_factors.is_empty() {
            let v = self.remaining_factors.iter().cloned().map(|v| {
                FactorType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("remainingFactors", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgVerifyChallengeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "verified",
            "session_id",
            "sessionId",
            "session_expires_at",
            "sessionExpiresAt",
            "remaining_factors",
            "remainingFactors",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Verified,
            SessionId,
            SessionExpiresAt,
            RemainingFactors,
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
                            "verified" => Ok(GeneratedField::Verified),
                            "sessionId" | "session_id" => Ok(GeneratedField::SessionId),
                            "sessionExpiresAt" | "session_expires_at" => Ok(GeneratedField::SessionExpiresAt),
                            "remainingFactors" | "remaining_factors" => Ok(GeneratedField::RemainingFactors),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgVerifyChallengeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.MsgVerifyChallengeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgVerifyChallengeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut verified__ = None;
                let mut session_id__ = None;
                let mut session_expires_at__ = None;
                let mut remaining_factors__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Verified => {
                            if verified__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verified"));
                            }
                            verified__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SessionId => {
                            if session_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionId"));
                            }
                            session_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SessionExpiresAt => {
                            if session_expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionExpiresAt"));
                            }
                            session_expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RemainingFactors => {
                            if remaining_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("remainingFactors"));
                            }
                            remaining_factors__ = Some(map_.next_value::<Vec<FactorType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                    }
                }
                Ok(MsgVerifyChallengeResponse {
                    verified: verified__.unwrap_or_default(),
                    session_id: session_id__.unwrap_or_default(),
                    session_expires_at: session_expires_at__.unwrap_or_default(),
                    remaining_factors: remaining_factors__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.MsgVerifyChallengeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for OtpChallengeInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delivery_method.is_empty() {
            len += 1;
        }
        if !self.delivery_destination_masked.is_empty() {
            len += 1;
        }
        if self.sent_at != 0 {
            len += 1;
        }
        if self.resend_count != 0 {
            len += 1;
        }
        if self.last_resend_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.OTPChallengeInfo", len)?;
        if !self.delivery_method.is_empty() {
            struct_ser.serialize_field("deliveryMethod", &self.delivery_method)?;
        }
        if !self.delivery_destination_masked.is_empty() {
            struct_ser.serialize_field("deliveryDestinationMasked", &self.delivery_destination_masked)?;
        }
        if self.sent_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sentAt", ToString::to_string(&self.sent_at).as_str())?;
        }
        if self.resend_count != 0 {
            struct_ser.serialize_field("resendCount", &self.resend_count)?;
        }
        if self.last_resend_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastResendAt", ToString::to_string(&self.last_resend_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for OtpChallengeInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delivery_method",
            "deliveryMethod",
            "delivery_destination_masked",
            "deliveryDestinationMasked",
            "sent_at",
            "sentAt",
            "resend_count",
            "resendCount",
            "last_resend_at",
            "lastResendAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DeliveryMethod,
            DeliveryDestinationMasked,
            SentAt,
            ResendCount,
            LastResendAt,
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
                            "deliveryMethod" | "delivery_method" => Ok(GeneratedField::DeliveryMethod),
                            "deliveryDestinationMasked" | "delivery_destination_masked" => Ok(GeneratedField::DeliveryDestinationMasked),
                            "sentAt" | "sent_at" => Ok(GeneratedField::SentAt),
                            "resendCount" | "resend_count" => Ok(GeneratedField::ResendCount),
                            "lastResendAt" | "last_resend_at" => Ok(GeneratedField::LastResendAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = OtpChallengeInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.OTPChallengeInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<OtpChallengeInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delivery_method__ = None;
                let mut delivery_destination_masked__ = None;
                let mut sent_at__ = None;
                let mut resend_count__ = None;
                let mut last_resend_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DeliveryMethod => {
                            if delivery_method__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deliveryMethod"));
                            }
                            delivery_method__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DeliveryDestinationMasked => {
                            if delivery_destination_masked__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deliveryDestinationMasked"));
                            }
                            delivery_destination_masked__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SentAt => {
                            if sent_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sentAt"));
                            }
                            sent_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ResendCount => {
                            if resend_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resendCount"));
                            }
                            resend_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastResendAt => {
                            if last_resend_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastResendAt"));
                            }
                            last_resend_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(OtpChallengeInfo {
                    delivery_method: delivery_method__.unwrap_or_default(),
                    delivery_destination_masked: delivery_destination_masked__.unwrap_or_default(),
                    sent_at: sent_at__.unwrap_or_default(),
                    resend_count: resend_count__.unwrap_or_default(),
                    last_resend_at: last_resend_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.OTPChallengeInfo", FIELDS, GeneratedVisitor)
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
        if self.default_session_duration != 0 {
            len += 1;
        }
        if self.max_factors_per_account != 0 {
            len += 1;
        }
        if self.max_challenge_attempts != 0 {
            len += 1;
        }
        if self.challenge_ttl != 0 {
            len += 1;
        }
        if self.max_trusted_devices != 0 {
            len += 1;
        }
        if self.trusted_device_ttl != 0 {
            len += 1;
        }
        if self.min_veid_score_for_mfa != 0 {
            len += 1;
        }
        if self.require_at_least_one_factor {
            len += 1;
        }
        if !self.allowed_factor_types.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.Params", len)?;
        if self.default_session_duration != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("defaultSessionDuration", ToString::to_string(&self.default_session_duration).as_str())?;
        }
        if self.max_factors_per_account != 0 {
            struct_ser.serialize_field("maxFactorsPerAccount", &self.max_factors_per_account)?;
        }
        if self.max_challenge_attempts != 0 {
            struct_ser.serialize_field("maxChallengeAttempts", &self.max_challenge_attempts)?;
        }
        if self.challenge_ttl != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challengeTtl", ToString::to_string(&self.challenge_ttl).as_str())?;
        }
        if self.max_trusted_devices != 0 {
            struct_ser.serialize_field("maxTrustedDevices", &self.max_trusted_devices)?;
        }
        if self.trusted_device_ttl != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("trustedDeviceTtl", ToString::to_string(&self.trusted_device_ttl).as_str())?;
        }
        if self.min_veid_score_for_mfa != 0 {
            struct_ser.serialize_field("minVeidScoreForMfa", &self.min_veid_score_for_mfa)?;
        }
        if self.require_at_least_one_factor {
            struct_ser.serialize_field("requireAtLeastOneFactor", &self.require_at_least_one_factor)?;
        }
        if !self.allowed_factor_types.is_empty() {
            let v = self.allowed_factor_types.iter().cloned().map(|v| {
                FactorType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("allowedFactorTypes", &v)?;
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
            "default_session_duration",
            "defaultSessionDuration",
            "max_factors_per_account",
            "maxFactorsPerAccount",
            "max_challenge_attempts",
            "maxChallengeAttempts",
            "challenge_ttl",
            "challengeTtl",
            "max_trusted_devices",
            "maxTrustedDevices",
            "trusted_device_ttl",
            "trustedDeviceTtl",
            "min_veid_score_for_mfa",
            "minVeidScoreForMfa",
            "require_at_least_one_factor",
            "requireAtLeastOneFactor",
            "allowed_factor_types",
            "allowedFactorTypes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DefaultSessionDuration,
            MaxFactorsPerAccount,
            MaxChallengeAttempts,
            ChallengeTtl,
            MaxTrustedDevices,
            TrustedDeviceTtl,
            MinVeidScoreForMfa,
            RequireAtLeastOneFactor,
            AllowedFactorTypes,
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
                            "defaultSessionDuration" | "default_session_duration" => Ok(GeneratedField::DefaultSessionDuration),
                            "maxFactorsPerAccount" | "max_factors_per_account" => Ok(GeneratedField::MaxFactorsPerAccount),
                            "maxChallengeAttempts" | "max_challenge_attempts" => Ok(GeneratedField::MaxChallengeAttempts),
                            "challengeTtl" | "challenge_ttl" => Ok(GeneratedField::ChallengeTtl),
                            "maxTrustedDevices" | "max_trusted_devices" => Ok(GeneratedField::MaxTrustedDevices),
                            "trustedDeviceTtl" | "trusted_device_ttl" => Ok(GeneratedField::TrustedDeviceTtl),
                            "minVeidScoreForMfa" | "min_veid_score_for_mfa" => Ok(GeneratedField::MinVeidScoreForMfa),
                            "requireAtLeastOneFactor" | "require_at_least_one_factor" => Ok(GeneratedField::RequireAtLeastOneFactor),
                            "allowedFactorTypes" | "allowed_factor_types" => Ok(GeneratedField::AllowedFactorTypes),
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
                formatter.write_str("struct virtengine.mfa.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut default_session_duration__ = None;
                let mut max_factors_per_account__ = None;
                let mut max_challenge_attempts__ = None;
                let mut challenge_ttl__ = None;
                let mut max_trusted_devices__ = None;
                let mut trusted_device_ttl__ = None;
                let mut min_veid_score_for_mfa__ = None;
                let mut require_at_least_one_factor__ = None;
                let mut allowed_factor_types__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DefaultSessionDuration => {
                            if default_session_duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("defaultSessionDuration"));
                            }
                            default_session_duration__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxFactorsPerAccount => {
                            if max_factors_per_account__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxFactorsPerAccount"));
                            }
                            max_factors_per_account__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxChallengeAttempts => {
                            if max_challenge_attempts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxChallengeAttempts"));
                            }
                            max_challenge_attempts__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ChallengeTtl => {
                            if challenge_ttl__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeTtl"));
                            }
                            challenge_ttl__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxTrustedDevices => {
                            if max_trusted_devices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxTrustedDevices"));
                            }
                            max_trusted_devices__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TrustedDeviceTtl => {
                            if trusted_device_ttl__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustedDeviceTtl"));
                            }
                            trusted_device_ttl__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinVeidScoreForMfa => {
                            if min_veid_score_for_mfa__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minVeidScoreForMfa"));
                            }
                            min_veid_score_for_mfa__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequireAtLeastOneFactor => {
                            if require_at_least_one_factor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireAtLeastOneFactor"));
                            }
                            require_at_least_one_factor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowedFactorTypes => {
                            if allowed_factor_types__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowedFactorTypes"));
                            }
                            allowed_factor_types__ = Some(map_.next_value::<Vec<FactorType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                    }
                }
                Ok(Params {
                    default_session_duration: default_session_duration__.unwrap_or_default(),
                    max_factors_per_account: max_factors_per_account__.unwrap_or_default(),
                    max_challenge_attempts: max_challenge_attempts__.unwrap_or_default(),
                    challenge_ttl: challenge_ttl__.unwrap_or_default(),
                    max_trusted_devices: max_trusted_devices__.unwrap_or_default(),
                    trusted_device_ttl: trusted_device_ttl__.unwrap_or_default(),
                    min_veid_score_for_mfa: min_veid_score_for_mfa__.unwrap_or_default(),
                    require_at_least_one_factor: require_at_least_one_factor__.unwrap_or_default(),
                    allowed_factor_types: allowed_factor_types__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAllSensitiveTxConfigsRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryAllSensitiveTxConfigsRequest", len)?;
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAllSensitiveTxConfigsRequest {
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
            type Value = QueryAllSensitiveTxConfigsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryAllSensitiveTxConfigsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAllSensitiveTxConfigsRequest, V::Error>
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
                Ok(QueryAllSensitiveTxConfigsRequest {
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryAllSensitiveTxConfigsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAllSensitiveTxConfigsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.configs.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryAllSensitiveTxConfigsResponse", len)?;
        if !self.configs.is_empty() {
            struct_ser.serialize_field("configs", &self.configs)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAllSensitiveTxConfigsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "configs",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Configs,
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
                            "configs" => Ok(GeneratedField::Configs),
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
            type Value = QueryAllSensitiveTxConfigsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryAllSensitiveTxConfigsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAllSensitiveTxConfigsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut configs__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Configs => {
                            if configs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("configs"));
                            }
                            configs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAllSensitiveTxConfigsResponse {
                    configs: configs__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryAllSensitiveTxConfigsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAuthorizationSessionRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.session_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryAuthorizationSessionRequest", len)?;
        if !self.session_id.is_empty() {
            struct_ser.serialize_field("sessionId", &self.session_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAuthorizationSessionRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "session_id",
            "sessionId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SessionId,
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
                            "sessionId" | "session_id" => Ok(GeneratedField::SessionId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAuthorizationSessionRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryAuthorizationSessionRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAuthorizationSessionRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut session_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SessionId => {
                            if session_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionId"));
                            }
                            session_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAuthorizationSessionRequest {
                    session_id: session_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryAuthorizationSessionRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAuthorizationSessionResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.session.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        if self.is_valid {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryAuthorizationSessionResponse", len)?;
        if let Some(v) = self.session.as_ref() {
            struct_ser.serialize_field("session", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        if self.is_valid {
            struct_ser.serialize_field("isValid", &self.is_valid)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAuthorizationSessionResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "session",
            "found",
            "is_valid",
            "isValid",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Session,
            Found,
            IsValid,
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
                            "session" => Ok(GeneratedField::Session),
                            "found" => Ok(GeneratedField::Found),
                            "isValid" | "is_valid" => Ok(GeneratedField::IsValid),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAuthorizationSessionResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryAuthorizationSessionResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAuthorizationSessionResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut session__ = None;
                let mut found__ = None;
                let mut is_valid__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Session => {
                            if session__.is_some() {
                                return Err(serde::de::Error::duplicate_field("session"));
                            }
                            session__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsValid => {
                            if is_valid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isValid"));
                            }
                            is_valid__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAuthorizationSessionResponse {
                    session: session__,
                    found: found__.unwrap_or_default(),
                    is_valid: is_valid__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryAuthorizationSessionResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryChallengeRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryChallengeRequest", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryChallengeRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryChallengeRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryChallengeRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryChallengeRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryChallengeRequest {
                    challenge_id: challenge_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryChallengeRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryChallengeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.challenge.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryChallengeResponse", len)?;
        if let Some(v) = self.challenge.as_ref() {
            struct_ser.serialize_field("challenge", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryChallengeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Challenge,
            Found,
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
                            "challenge" => Ok(GeneratedField::Challenge),
                            "found" => Ok(GeneratedField::Found),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryChallengeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryChallengeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryChallengeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Challenge => {
                            if challenge__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challenge"));
                            }
                            challenge__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryChallengeResponse {
                    challenge: challenge__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryChallengeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFactorEnrollmentRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.address.is_empty() {
            len += 1;
        }
        if !self.factor_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.factor_id.is_empty() {
            struct_ser.serialize_field("factorId", &self.factor_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFactorEnrollmentRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "factor_id",
            "factorId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            FactorId,
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
                            "address" => Ok(GeneratedField::Address),
                            "factorId" | "factor_id" => Ok(GeneratedField::FactorId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFactorEnrollmentRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryFactorEnrollmentRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFactorEnrollmentRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut factor_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorId => {
                            if factor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorId"));
                            }
                            factor_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryFactorEnrollmentRequest {
                    address: address__.unwrap_or_default(),
                    factor_id: factor_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFactorEnrollmentResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.enrollment.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentResponse", len)?;
        if let Some(v) = self.enrollment.as_ref() {
            struct_ser.serialize_field("enrollment", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFactorEnrollmentResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "enrollment",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Enrollment,
            Found,
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
                            "enrollment" => Ok(GeneratedField::Enrollment),
                            "found" => Ok(GeneratedField::Found),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryFactorEnrollmentResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryFactorEnrollmentResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFactorEnrollmentResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut enrollment__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Enrollment => {
                            if enrollment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enrollment"));
                            }
                            enrollment__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryFactorEnrollmentResponse {
                    enrollment: enrollment__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFactorEnrollmentsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.address.is_empty() {
            len += 1;
        }
        if self.factor_type_filter != 0 {
            len += 1;
        }
        if self.status_filter != 0 {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentsRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if self.factor_type_filter != 0 {
            let v = FactorType::try_from(self.factor_type_filter)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.factor_type_filter)))?;
            struct_ser.serialize_field("factorTypeFilter", &v)?;
        }
        if self.status_filter != 0 {
            let v = FactorEnrollmentStatus::try_from(self.status_filter)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status_filter)))?;
            struct_ser.serialize_field("statusFilter", &v)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFactorEnrollmentsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "factor_type_filter",
            "factorTypeFilter",
            "status_filter",
            "statusFilter",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            FactorTypeFilter,
            StatusFilter,
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
                            "address" => Ok(GeneratedField::Address),
                            "factorTypeFilter" | "factor_type_filter" => Ok(GeneratedField::FactorTypeFilter),
                            "statusFilter" | "status_filter" => Ok(GeneratedField::StatusFilter),
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
            type Value = QueryFactorEnrollmentsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryFactorEnrollmentsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFactorEnrollmentsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut factor_type_filter__ = None;
                let mut status_filter__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorTypeFilter => {
                            if factor_type_filter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorTypeFilter"));
                            }
                            factor_type_filter__ = Some(map_.next_value::<FactorType>()? as i32);
                        }
                        GeneratedField::StatusFilter => {
                            if status_filter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("statusFilter"));
                            }
                            status_filter__ = Some(map_.next_value::<FactorEnrollmentStatus>()? as i32);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFactorEnrollmentsRequest {
                    address: address__.unwrap_or_default(),
                    factor_type_filter: factor_type_filter__.unwrap_or_default(),
                    status_filter: status_filter__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryFactorEnrollmentsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.enrollments.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentsResponse", len)?;
        if !self.enrollments.is_empty() {
            struct_ser.serialize_field("enrollments", &self.enrollments)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryFactorEnrollmentsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "enrollments",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Enrollments,
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
                            "enrollments" => Ok(GeneratedField::Enrollments),
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
            type Value = QueryFactorEnrollmentsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryFactorEnrollmentsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryFactorEnrollmentsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut enrollments__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Enrollments => {
                            if enrollments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enrollments"));
                            }
                            enrollments__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryFactorEnrollmentsResponse {
                    enrollments: enrollments__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryFactorEnrollmentsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMfaPolicyRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryMFAPolicyRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMfaPolicyRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
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
                            "address" => Ok(GeneratedField::Address),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryMfaPolicyRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryMFAPolicyRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMfaPolicyRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryMfaPolicyRequest {
                    address: address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryMFAPolicyRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMfaPolicyResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.policy.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryMFAPolicyResponse", len)?;
        if let Some(v) = self.policy.as_ref() {
            struct_ser.serialize_field("policy", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMfaPolicyResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "policy",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Policy,
            Found,
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
                            "policy" => Ok(GeneratedField::Policy),
                            "found" => Ok(GeneratedField::Found),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryMfaPolicyResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryMFAPolicyResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMfaPolicyResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut policy__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Policy => {
                            if policy__.is_some() {
                                return Err(serde::de::Error::duplicate_field("policy"));
                            }
                            policy__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryMfaPolicyResponse {
                    policy: policy__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryMFAPolicyResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMfaRequiredRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.address.is_empty() {
            len += 1;
        }
        if self.transaction_type != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryMFARequiredRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if self.transaction_type != 0 {
            let v = SensitiveTransactionType::try_from(self.transaction_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.transaction_type)))?;
            struct_ser.serialize_field("transactionType", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMfaRequiredRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "transaction_type",
            "transactionType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            TransactionType,
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
                            "address" => Ok(GeneratedField::Address),
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryMfaRequiredRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryMFARequiredRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMfaRequiredRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut transaction_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value::<SensitiveTransactionType>()? as i32);
                        }
                    }
                }
                Ok(QueryMfaRequiredRequest {
                    address: address__.unwrap_or_default(),
                    transaction_type: transaction_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryMFARequiredRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryMfaRequiredResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.required {
            len += 1;
        }
        if !self.factor_combinations.is_empty() {
            len += 1;
        }
        if self.min_veid_score != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryMFARequiredResponse", len)?;
        if self.required {
            struct_ser.serialize_field("required", &self.required)?;
        }
        if !self.factor_combinations.is_empty() {
            struct_ser.serialize_field("factorCombinations", &self.factor_combinations)?;
        }
        if self.min_veid_score != 0 {
            struct_ser.serialize_field("minVeidScore", &self.min_veid_score)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryMfaRequiredResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "required",
            "factor_combinations",
            "factorCombinations",
            "min_veid_score",
            "minVeidScore",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Required,
            FactorCombinations,
            MinVeidScore,
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
                            "required" => Ok(GeneratedField::Required),
                            "factorCombinations" | "factor_combinations" => Ok(GeneratedField::FactorCombinations),
                            "minVeidScore" | "min_veid_score" => Ok(GeneratedField::MinVeidScore),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryMfaRequiredResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryMFARequiredResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryMfaRequiredResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut required__ = None;
                let mut factor_combinations__ = None;
                let mut min_veid_score__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Required => {
                            if required__.is_some() {
                                return Err(serde::de::Error::duplicate_field("required"));
                            }
                            required__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorCombinations => {
                            if factor_combinations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorCombinations"));
                            }
                            factor_combinations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinVeidScore => {
                            if min_veid_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minVeidScore"));
                            }
                            min_veid_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryMfaRequiredResponse {
                    required: required__.unwrap_or_default(),
                    factor_combinations: factor_combinations__.unwrap_or_default(),
                    min_veid_score: min_veid_score__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryMFARequiredResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.mfa.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.mfa.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPendingChallengesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryPendingChallengesRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPendingChallengesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
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
                            "address" => Ok(GeneratedField::Address),
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
            type Value = QueryPendingChallengesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryPendingChallengesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPendingChallengesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryPendingChallengesRequest {
                    address: address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryPendingChallengesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPendingChallengesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenges.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryPendingChallengesResponse", len)?;
        if !self.challenges.is_empty() {
            struct_ser.serialize_field("challenges", &self.challenges)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPendingChallengesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenges",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Challenges,
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
                            "challenges" => Ok(GeneratedField::Challenges),
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
            type Value = QueryPendingChallengesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryPendingChallengesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPendingChallengesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenges__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Challenges => {
                            if challenges__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challenges"));
                            }
                            challenges__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryPendingChallengesResponse {
                    challenges: challenges__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryPendingChallengesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuerySensitiveTxConfigRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.transaction_type != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QuerySensitiveTxConfigRequest", len)?;
        if self.transaction_type != 0 {
            let v = SensitiveTransactionType::try_from(self.transaction_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.transaction_type)))?;
            struct_ser.serialize_field("transactionType", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuerySensitiveTxConfigRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "transaction_type",
            "transactionType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TransactionType,
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
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuerySensitiveTxConfigRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QuerySensitiveTxConfigRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuerySensitiveTxConfigRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut transaction_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value::<SensitiveTransactionType>()? as i32);
                        }
                    }
                }
                Ok(QuerySensitiveTxConfigRequest {
                    transaction_type: transaction_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QuerySensitiveTxConfigRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuerySensitiveTxConfigResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.config.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QuerySensitiveTxConfigResponse", len)?;
        if let Some(v) = self.config.as_ref() {
            struct_ser.serialize_field("config", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuerySensitiveTxConfigResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "config",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Config,
            Found,
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
                            "config" => Ok(GeneratedField::Config),
                            "found" => Ok(GeneratedField::Found),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuerySensitiveTxConfigResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QuerySensitiveTxConfigResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuerySensitiveTxConfigResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut config__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Config => {
                            if config__.is_some() {
                                return Err(serde::de::Error::duplicate_field("config"));
                            }
                            config__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QuerySensitiveTxConfigResponse {
                    config: config__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QuerySensitiveTxConfigResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryTrustedDevicesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryTrustedDevicesRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryTrustedDevicesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
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
                            "address" => Ok(GeneratedField::Address),
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
            type Value = QueryTrustedDevicesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryTrustedDevicesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryTrustedDevicesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryTrustedDevicesRequest {
                    address: address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryTrustedDevicesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryTrustedDevicesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.devices.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.QueryTrustedDevicesResponse", len)?;
        if !self.devices.is_empty() {
            struct_ser.serialize_field("devices", &self.devices)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryTrustedDevicesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "devices",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Devices,
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
                            "devices" => Ok(GeneratedField::Devices),
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
            type Value = QueryTrustedDevicesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.QueryTrustedDevicesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryTrustedDevicesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut devices__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Devices => {
                            if devices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("devices"));
                            }
                            devices__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryTrustedDevicesResponse {
                    devices: devices__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.QueryTrustedDevicesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RevocationStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unknown => "REVOCATION_STATUS_UNKNOWN",
            Self::Good => "REVOCATION_STATUS_GOOD",
            Self::Revoked => "REVOCATION_STATUS_REVOKED",
            Self::CheckFailed => "REVOCATION_STATUS_CHECK_FAILED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for RevocationStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "REVOCATION_STATUS_UNKNOWN",
            "REVOCATION_STATUS_GOOD",
            "REVOCATION_STATUS_REVOKED",
            "REVOCATION_STATUS_CHECK_FAILED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RevocationStatus;

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
                    "REVOCATION_STATUS_UNKNOWN" => Ok(RevocationStatus::Unknown),
                    "REVOCATION_STATUS_GOOD" => Ok(RevocationStatus::Good),
                    "REVOCATION_STATUS_REVOKED" => Ok(RevocationStatus::Revoked),
                    "REVOCATION_STATUS_CHECK_FAILED" => Ok(RevocationStatus::CheckFailed),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for SensitiveTransactionType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::SensitiveTxUnspecified => "SENSITIVE_TX_UNSPECIFIED",
            Self::SensitiveTxAccountRecovery => "SENSITIVE_TX_ACCOUNT_RECOVERY",
            Self::SensitiveTxKeyRotation => "SENSITIVE_TX_KEY_ROTATION",
            Self::SensitiveTxLargeWithdrawal => "SENSITIVE_TX_LARGE_WITHDRAWAL",
            Self::SensitiveTxProviderRegistration => "SENSITIVE_TX_PROVIDER_REGISTRATION",
            Self::SensitiveTxValidatorRegistration => "SENSITIVE_TX_VALIDATOR_REGISTRATION",
            Self::SensitiveTxHighValueOrder => "SENSITIVE_TX_HIGH_VALUE_ORDER",
            Self::SensitiveTxRoleAssignment => "SENSITIVE_TX_ROLE_ASSIGNMENT",
            Self::SensitiveTxGovernanceProposal => "SENSITIVE_TX_GOVERNANCE_PROPOSAL",
            Self::SensitiveTxPrimaryEmailChange => "SENSITIVE_TX_PRIMARY_EMAIL_CHANGE",
            Self::SensitiveTxPhoneNumberChange => "SENSITIVE_TX_PHONE_NUMBER_CHANGE",
            Self::SensitiveTxTwoFactorDisable => "SENSITIVE_TX_TWO_FACTOR_DISABLE",
            Self::SensitiveTxAccountDeletion => "SENSITIVE_TX_ACCOUNT_DELETION",
            Self::SensitiveTxGovernanceVote => "SENSITIVE_TX_GOVERNANCE_VOTE",
            Self::SensitiveTxFirstOfferingCreate => "SENSITIVE_TX_FIRST_OFFERING_CREATE",
            Self::SensitiveTxTransferToNewAddress => "SENSITIVE_TX_TRANSFER_TO_NEW_ADDRESS",
            Self::SensitiveTxMediumWithdrawal => "SENSITIVE_TX_MEDIUM_WITHDRAWAL",
            Self::SensitiveTxApiKeyGeneration => "SENSITIVE_TX_API_KEY_GENERATION",
            Self::SensitiveTxWebhookConfiguration => "SENSITIVE_TX_WEBHOOK_CONFIGURATION",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for SensitiveTransactionType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "SENSITIVE_TX_UNSPECIFIED",
            "SENSITIVE_TX_ACCOUNT_RECOVERY",
            "SENSITIVE_TX_KEY_ROTATION",
            "SENSITIVE_TX_LARGE_WITHDRAWAL",
            "SENSITIVE_TX_PROVIDER_REGISTRATION",
            "SENSITIVE_TX_VALIDATOR_REGISTRATION",
            "SENSITIVE_TX_HIGH_VALUE_ORDER",
            "SENSITIVE_TX_ROLE_ASSIGNMENT",
            "SENSITIVE_TX_GOVERNANCE_PROPOSAL",
            "SENSITIVE_TX_PRIMARY_EMAIL_CHANGE",
            "SENSITIVE_TX_PHONE_NUMBER_CHANGE",
            "SENSITIVE_TX_TWO_FACTOR_DISABLE",
            "SENSITIVE_TX_ACCOUNT_DELETION",
            "SENSITIVE_TX_GOVERNANCE_VOTE",
            "SENSITIVE_TX_FIRST_OFFERING_CREATE",
            "SENSITIVE_TX_TRANSFER_TO_NEW_ADDRESS",
            "SENSITIVE_TX_MEDIUM_WITHDRAWAL",
            "SENSITIVE_TX_API_KEY_GENERATION",
            "SENSITIVE_TX_WEBHOOK_CONFIGURATION",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SensitiveTransactionType;

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
                    "SENSITIVE_TX_UNSPECIFIED" => Ok(SensitiveTransactionType::SensitiveTxUnspecified),
                    "SENSITIVE_TX_ACCOUNT_RECOVERY" => Ok(SensitiveTransactionType::SensitiveTxAccountRecovery),
                    "SENSITIVE_TX_KEY_ROTATION" => Ok(SensitiveTransactionType::SensitiveTxKeyRotation),
                    "SENSITIVE_TX_LARGE_WITHDRAWAL" => Ok(SensitiveTransactionType::SensitiveTxLargeWithdrawal),
                    "SENSITIVE_TX_PROVIDER_REGISTRATION" => Ok(SensitiveTransactionType::SensitiveTxProviderRegistration),
                    "SENSITIVE_TX_VALIDATOR_REGISTRATION" => Ok(SensitiveTransactionType::SensitiveTxValidatorRegistration),
                    "SENSITIVE_TX_HIGH_VALUE_ORDER" => Ok(SensitiveTransactionType::SensitiveTxHighValueOrder),
                    "SENSITIVE_TX_ROLE_ASSIGNMENT" => Ok(SensitiveTransactionType::SensitiveTxRoleAssignment),
                    "SENSITIVE_TX_GOVERNANCE_PROPOSAL" => Ok(SensitiveTransactionType::SensitiveTxGovernanceProposal),
                    "SENSITIVE_TX_PRIMARY_EMAIL_CHANGE" => Ok(SensitiveTransactionType::SensitiveTxPrimaryEmailChange),
                    "SENSITIVE_TX_PHONE_NUMBER_CHANGE" => Ok(SensitiveTransactionType::SensitiveTxPhoneNumberChange),
                    "SENSITIVE_TX_TWO_FACTOR_DISABLE" => Ok(SensitiveTransactionType::SensitiveTxTwoFactorDisable),
                    "SENSITIVE_TX_ACCOUNT_DELETION" => Ok(SensitiveTransactionType::SensitiveTxAccountDeletion),
                    "SENSITIVE_TX_GOVERNANCE_VOTE" => Ok(SensitiveTransactionType::SensitiveTxGovernanceVote),
                    "SENSITIVE_TX_FIRST_OFFERING_CREATE" => Ok(SensitiveTransactionType::SensitiveTxFirstOfferingCreate),
                    "SENSITIVE_TX_TRANSFER_TO_NEW_ADDRESS" => Ok(SensitiveTransactionType::SensitiveTxTransferToNewAddress),
                    "SENSITIVE_TX_MEDIUM_WITHDRAWAL" => Ok(SensitiveTransactionType::SensitiveTxMediumWithdrawal),
                    "SENSITIVE_TX_API_KEY_GENERATION" => Ok(SensitiveTransactionType::SensitiveTxApiKeyGeneration),
                    "SENSITIVE_TX_WEBHOOK_CONFIGURATION" => Ok(SensitiveTransactionType::SensitiveTxWebhookConfiguration),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for SensitiveTxConfig {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.transaction_type != 0 {
            len += 1;
        }
        if self.enabled {
            len += 1;
        }
        if self.min_veid_score != 0 {
            len += 1;
        }
        if !self.required_factor_combinations.is_empty() {
            len += 1;
        }
        if self.session_duration != 0 {
            len += 1;
        }
        if self.is_single_use {
            len += 1;
        }
        if self.allow_trusted_device_reduction {
            len += 1;
        }
        if !self.value_threshold.is_empty() {
            len += 1;
        }
        if self.cooldown_period != 0 {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.SensitiveTxConfig", len)?;
        if self.transaction_type != 0 {
            let v = SensitiveTransactionType::try_from(self.transaction_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.transaction_type)))?;
            struct_ser.serialize_field("transactionType", &v)?;
        }
        if self.enabled {
            struct_ser.serialize_field("enabled", &self.enabled)?;
        }
        if self.min_veid_score != 0 {
            struct_ser.serialize_field("minVeidScore", &self.min_veid_score)?;
        }
        if !self.required_factor_combinations.is_empty() {
            struct_ser.serialize_field("requiredFactorCombinations", &self.required_factor_combinations)?;
        }
        if self.session_duration != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sessionDuration", ToString::to_string(&self.session_duration).as_str())?;
        }
        if self.is_single_use {
            struct_ser.serialize_field("isSingleUse", &self.is_single_use)?;
        }
        if self.allow_trusted_device_reduction {
            struct_ser.serialize_field("allowTrustedDeviceReduction", &self.allow_trusted_device_reduction)?;
        }
        if !self.value_threshold.is_empty() {
            struct_ser.serialize_field("valueThreshold", &self.value_threshold)?;
        }
        if self.cooldown_period != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("cooldownPeriod", ToString::to_string(&self.cooldown_period).as_str())?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SensitiveTxConfig {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "transaction_type",
            "transactionType",
            "enabled",
            "min_veid_score",
            "minVeidScore",
            "required_factor_combinations",
            "requiredFactorCombinations",
            "session_duration",
            "sessionDuration",
            "is_single_use",
            "isSingleUse",
            "allow_trusted_device_reduction",
            "allowTrustedDeviceReduction",
            "value_threshold",
            "valueThreshold",
            "cooldown_period",
            "cooldownPeriod",
            "description",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TransactionType,
            Enabled,
            MinVeidScore,
            RequiredFactorCombinations,
            SessionDuration,
            IsSingleUse,
            AllowTrustedDeviceReduction,
            ValueThreshold,
            CooldownPeriod,
            Description,
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
                            "transactionType" | "transaction_type" => Ok(GeneratedField::TransactionType),
                            "enabled" => Ok(GeneratedField::Enabled),
                            "minVeidScore" | "min_veid_score" => Ok(GeneratedField::MinVeidScore),
                            "requiredFactorCombinations" | "required_factor_combinations" => Ok(GeneratedField::RequiredFactorCombinations),
                            "sessionDuration" | "session_duration" => Ok(GeneratedField::SessionDuration),
                            "isSingleUse" | "is_single_use" => Ok(GeneratedField::IsSingleUse),
                            "allowTrustedDeviceReduction" | "allow_trusted_device_reduction" => Ok(GeneratedField::AllowTrustedDeviceReduction),
                            "valueThreshold" | "value_threshold" => Ok(GeneratedField::ValueThreshold),
                            "cooldownPeriod" | "cooldown_period" => Ok(GeneratedField::CooldownPeriod),
                            "description" => Ok(GeneratedField::Description),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SensitiveTxConfig;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.SensitiveTxConfig")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SensitiveTxConfig, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut transaction_type__ = None;
                let mut enabled__ = None;
                let mut min_veid_score__ = None;
                let mut required_factor_combinations__ = None;
                let mut session_duration__ = None;
                let mut is_single_use__ = None;
                let mut allow_trusted_device_reduction__ = None;
                let mut value_threshold__ = None;
                let mut cooldown_period__ = None;
                let mut description__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TransactionType => {
                            if transaction_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transactionType"));
                            }
                            transaction_type__ = Some(map_.next_value::<SensitiveTransactionType>()? as i32);
                        }
                        GeneratedField::Enabled => {
                            if enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enabled"));
                            }
                            enabled__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinVeidScore => {
                            if min_veid_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minVeidScore"));
                            }
                            min_veid_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequiredFactorCombinations => {
                            if required_factor_combinations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredFactorCombinations"));
                            }
                            required_factor_combinations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SessionDuration => {
                            if session_duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sessionDuration"));
                            }
                            session_duration__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsSingleUse => {
                            if is_single_use__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isSingleUse"));
                            }
                            is_single_use__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowTrustedDeviceReduction => {
                            if allow_trusted_device_reduction__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowTrustedDeviceReduction"));
                            }
                            allow_trusted_device_reduction__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValueThreshold => {
                            if value_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("valueThreshold"));
                            }
                            value_threshold__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CooldownPeriod => {
                            if cooldown_period__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cooldownPeriod"));
                            }
                            cooldown_period__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(SensitiveTxConfig {
                    transaction_type: transaction_type__.unwrap_or_default(),
                    enabled: enabled__.unwrap_or_default(),
                    min_veid_score: min_veid_score__.unwrap_or_default(),
                    required_factor_combinations: required_factor_combinations__.unwrap_or_default(),
                    session_duration: session_duration__.unwrap_or_default(),
                    is_single_use: is_single_use__.unwrap_or_default(),
                    allow_trusted_device_reduction: allow_trusted_device_reduction__.unwrap_or_default(),
                    value_threshold: value_threshold__.unwrap_or_default(),
                    cooldown_period: cooldown_period__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.SensitiveTxConfig", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SmartCardInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.card_serial_number.is_empty() {
            len += 1;
        }
        if !self.card_type.is_empty() {
            len += 1;
        }
        if !self.slot_id.is_empty() {
            len += 1;
        }
        if !self.chuid.is_empty() {
            len += 1;
        }
        if !self.fascn.is_empty() {
            len += 1;
        }
        if !self.card_holder_name.is_empty() {
            len += 1;
        }
        if self.expiration_date != 0 {
            len += 1;
        }
        if self.last_pin_verification != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.SmartCardInfo", len)?;
        if !self.card_serial_number.is_empty() {
            struct_ser.serialize_field("cardSerialNumber", &self.card_serial_number)?;
        }
        if !self.card_type.is_empty() {
            struct_ser.serialize_field("cardType", &self.card_type)?;
        }
        if !self.slot_id.is_empty() {
            struct_ser.serialize_field("slotId", &self.slot_id)?;
        }
        if !self.chuid.is_empty() {
            struct_ser.serialize_field("chuid", &self.chuid)?;
        }
        if !self.fascn.is_empty() {
            struct_ser.serialize_field("fascn", &self.fascn)?;
        }
        if !self.card_holder_name.is_empty() {
            struct_ser.serialize_field("cardHolderName", &self.card_holder_name)?;
        }
        if self.expiration_date != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expirationDate", ToString::to_string(&self.expiration_date).as_str())?;
        }
        if self.last_pin_verification != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastPinVerification", ToString::to_string(&self.last_pin_verification).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SmartCardInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "card_serial_number",
            "cardSerialNumber",
            "card_type",
            "cardType",
            "slot_id",
            "slotId",
            "chuid",
            "fascn",
            "card_holder_name",
            "cardHolderName",
            "expiration_date",
            "expirationDate",
            "last_pin_verification",
            "lastPinVerification",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CardSerialNumber,
            CardType,
            SlotId,
            Chuid,
            Fascn,
            CardHolderName,
            ExpirationDate,
            LastPinVerification,
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
                            "cardSerialNumber" | "card_serial_number" => Ok(GeneratedField::CardSerialNumber),
                            "cardType" | "card_type" => Ok(GeneratedField::CardType),
                            "slotId" | "slot_id" => Ok(GeneratedField::SlotId),
                            "chuid" => Ok(GeneratedField::Chuid),
                            "fascn" => Ok(GeneratedField::Fascn),
                            "cardHolderName" | "card_holder_name" => Ok(GeneratedField::CardHolderName),
                            "expirationDate" | "expiration_date" => Ok(GeneratedField::ExpirationDate),
                            "lastPinVerification" | "last_pin_verification" => Ok(GeneratedField::LastPinVerification),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SmartCardInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.SmartCardInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SmartCardInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut card_serial_number__ = None;
                let mut card_type__ = None;
                let mut slot_id__ = None;
                let mut chuid__ = None;
                let mut fascn__ = None;
                let mut card_holder_name__ = None;
                let mut expiration_date__ = None;
                let mut last_pin_verification__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CardSerialNumber => {
                            if card_serial_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cardSerialNumber"));
                            }
                            card_serial_number__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CardType => {
                            if card_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cardType"));
                            }
                            card_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SlotId => {
                            if slot_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slotId"));
                            }
                            slot_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Chuid => {
                            if chuid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("chuid"));
                            }
                            chuid__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Fascn => {
                            if fascn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fascn"));
                            }
                            fascn__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CardHolderName => {
                            if card_holder_name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cardHolderName"));
                            }
                            card_holder_name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpirationDate => {
                            if expiration_date__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expirationDate"));
                            }
                            expiration_date__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastPinVerification => {
                            if last_pin_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastPinVerification"));
                            }
                            last_pin_verification__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(SmartCardInfo {
                    card_serial_number: card_serial_number__.unwrap_or_default(),
                    card_type: card_type__.unwrap_or_default(),
                    slot_id: slot_id__.unwrap_or_default(),
                    chuid: chuid__.unwrap_or_default(),
                    fascn: fascn__.unwrap_or_default(),
                    card_holder_name: card_holder_name__.unwrap_or_default(),
                    expiration_date: expiration_date__.unwrap_or_default(),
                    last_pin_verification: last_pin_verification__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.SmartCardInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for TrustedDevice {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.device_info.is_some() {
            len += 1;
        }
        if self.added_at != 0 {
            len += 1;
        }
        if self.last_used_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.TrustedDevice", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if let Some(v) = self.device_info.as_ref() {
            struct_ser.serialize_field("deviceInfo", v)?;
        }
        if self.added_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("addedAt", ToString::to_string(&self.added_at).as_str())?;
        }
        if self.last_used_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUsedAt", ToString::to_string(&self.last_used_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for TrustedDevice {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "device_info",
            "deviceInfo",
            "added_at",
            "addedAt",
            "last_used_at",
            "lastUsedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            DeviceInfo,
            AddedAt,
            LastUsedAt,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "deviceInfo" | "device_info" => Ok(GeneratedField::DeviceInfo),
                            "addedAt" | "added_at" => Ok(GeneratedField::AddedAt),
                            "lastUsedAt" | "last_used_at" => Ok(GeneratedField::LastUsedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = TrustedDevice;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.TrustedDevice")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<TrustedDevice, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut device_info__ = None;
                let mut added_at__ = None;
                let mut last_used_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DeviceInfo => {
                            if device_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceInfo"));
                            }
                            device_info__ = map_.next_value()?;
                        }
                        GeneratedField::AddedAt => {
                            if added_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addedAt"));
                            }
                            added_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastUsedAt => {
                            if last_used_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUsedAt"));
                            }
                            last_used_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(TrustedDevice {
                    account_address: account_address__.unwrap_or_default(),
                    device_info: device_info__,
                    added_at: added_at__.unwrap_or_default(),
                    last_used_at: last_used_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.TrustedDevice", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for TrustedDevicePolicy {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.enabled {
            len += 1;
        }
        if self.trust_duration != 0 {
            len += 1;
        }
        if self.reduced_factors.is_some() {
            len += 1;
        }
        if self.max_trusted_devices != 0 {
            len += 1;
        }
        if self.require_reauth_for_sensitive {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.mfa.v1.TrustedDevicePolicy", len)?;
        if self.enabled {
            struct_ser.serialize_field("enabled", &self.enabled)?;
        }
        if self.trust_duration != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("trustDuration", ToString::to_string(&self.trust_duration).as_str())?;
        }
        if let Some(v) = self.reduced_factors.as_ref() {
            struct_ser.serialize_field("reducedFactors", v)?;
        }
        if self.max_trusted_devices != 0 {
            struct_ser.serialize_field("maxTrustedDevices", &self.max_trusted_devices)?;
        }
        if self.require_reauth_for_sensitive {
            struct_ser.serialize_field("requireReauthForSensitive", &self.require_reauth_for_sensitive)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for TrustedDevicePolicy {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "enabled",
            "trust_duration",
            "trustDuration",
            "reduced_factors",
            "reducedFactors",
            "max_trusted_devices",
            "maxTrustedDevices",
            "require_reauth_for_sensitive",
            "requireReauthForSensitive",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Enabled,
            TrustDuration,
            ReducedFactors,
            MaxTrustedDevices,
            RequireReauthForSensitive,
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
                            "enabled" => Ok(GeneratedField::Enabled),
                            "trustDuration" | "trust_duration" => Ok(GeneratedField::TrustDuration),
                            "reducedFactors" | "reduced_factors" => Ok(GeneratedField::ReducedFactors),
                            "maxTrustedDevices" | "max_trusted_devices" => Ok(GeneratedField::MaxTrustedDevices),
                            "requireReauthForSensitive" | "require_reauth_for_sensitive" => Ok(GeneratedField::RequireReauthForSensitive),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = TrustedDevicePolicy;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.mfa.v1.TrustedDevicePolicy")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<TrustedDevicePolicy, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut enabled__ = None;
                let mut trust_duration__ = None;
                let mut reduced_factors__ = None;
                let mut max_trusted_devices__ = None;
                let mut require_reauth_for_sensitive__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Enabled => {
                            if enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enabled"));
                            }
                            enabled__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TrustDuration => {
                            if trust_duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustDuration"));
                            }
                            trust_duration__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReducedFactors => {
                            if reduced_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reducedFactors"));
                            }
                            reduced_factors__ = map_.next_value()?;
                        }
                        GeneratedField::MaxTrustedDevices => {
                            if max_trusted_devices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxTrustedDevices"));
                            }
                            max_trusted_devices__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequireReauthForSensitive => {
                            if require_reauth_for_sensitive__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireReauthForSensitive"));
                            }
                            require_reauth_for_sensitive__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(TrustedDevicePolicy {
                    enabled: enabled__.unwrap_or_default(),
                    trust_duration: trust_duration__.unwrap_or_default(),
                    reduced_factors: reduced_factors__,
                    max_trusted_devices: max_trusted_devices__.unwrap_or_default(),
                    require_reauth_for_sensitive: require_reauth_for_sensitive__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.mfa.v1.TrustedDevicePolicy", FIELDS, GeneratedVisitor)
    }
}

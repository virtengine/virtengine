// @generated
impl serde::Serialize for AlgorithmInfo {
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
        if self.version != 0 {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.key_size != 0 {
            len += 1;
        }
        if self.nonce_size != 0 {
            len += 1;
        }
        if self.deprecated {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.AlgorithmInfo", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if self.version != 0 {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.key_size != 0 {
            struct_ser.serialize_field("keySize", &self.key_size)?;
        }
        if self.nonce_size != 0 {
            struct_ser.serialize_field("nonceSize", &self.nonce_size)?;
        }
        if self.deprecated {
            struct_ser.serialize_field("deprecated", &self.deprecated)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AlgorithmInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "version",
            "description",
            "key_size",
            "keySize",
            "nonce_size",
            "nonceSize",
            "deprecated",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Version,
            Description,
            KeySize,
            NonceSize,
            Deprecated,
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
                            "version" => Ok(GeneratedField::Version),
                            "description" => Ok(GeneratedField::Description),
                            "keySize" | "key_size" => Ok(GeneratedField::KeySize),
                            "nonceSize" | "nonce_size" => Ok(GeneratedField::NonceSize),
                            "deprecated" => Ok(GeneratedField::Deprecated),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AlgorithmInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.AlgorithmInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AlgorithmInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut version__ = None;
                let mut description__ = None;
                let mut key_size__ = None;
                let mut nonce_size__ = None;
                let mut deprecated__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::KeySize => {
                            if key_size__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keySize"));
                            }
                            key_size__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NonceSize => {
                            if nonce_size__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nonceSize"));
                            }
                            nonce_size__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Deprecated => {
                            if deprecated__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deprecated"));
                            }
                            deprecated__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(AlgorithmInfo {
                    id: id__.unwrap_or_default(),
                    version: version__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    key_size: key_size__.unwrap_or_default(),
                    nonce_size: nonce_size__.unwrap_or_default(),
                    deprecated: deprecated__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.AlgorithmInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EncryptedPayloadEnvelope {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.version != 0 {
            len += 1;
        }
        if !self.algorithm_id.is_empty() {
            len += 1;
        }
        if self.algorithm_version != 0 {
            len += 1;
        }
        if !self.recipient_key_ids.is_empty() {
            len += 1;
        }
        if !self.recipient_public_keys.is_empty() {
            len += 1;
        }
        if !self.encrypted_keys.is_empty() {
            len += 1;
        }
        if !self.wrapped_keys.is_empty() {
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
        if !self.metadata.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.EncryptedPayloadEnvelope", len)?;
        if self.version != 0 {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if !self.algorithm_id.is_empty() {
            struct_ser.serialize_field("algorithmId", &self.algorithm_id)?;
        }
        if self.algorithm_version != 0 {
            struct_ser.serialize_field("algorithmVersion", &self.algorithm_version)?;
        }
        if !self.recipient_key_ids.is_empty() {
            struct_ser.serialize_field("recipientKeyIds", &self.recipient_key_ids)?;
        }
        if !self.recipient_public_keys.is_empty() {
            struct_ser.serialize_field("recipientPublicKeys", &self.recipient_public_keys.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.encrypted_keys.is_empty() {
            struct_ser.serialize_field("encryptedKeys", &self.encrypted_keys.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.wrapped_keys.is_empty() {
            struct_ser.serialize_field("wrappedKeys", &self.wrapped_keys)?;
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
        if !self.metadata.is_empty() {
            struct_ser.serialize_field("metadata", &self.metadata)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EncryptedPayloadEnvelope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "version",
            "algorithm_id",
            "algorithmId",
            "algorithm_version",
            "algorithmVersion",
            "recipient_key_ids",
            "recipientKeyIds",
            "recipient_public_keys",
            "recipientPublicKeys",
            "encrypted_keys",
            "encryptedKeys",
            "wrapped_keys",
            "wrappedKeys",
            "nonce",
            "ciphertext",
            "sender_signature",
            "senderSignature",
            "sender_pub_key",
            "senderPubKey",
            "metadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Version,
            AlgorithmId,
            AlgorithmVersion,
            RecipientKeyIds,
            RecipientPublicKeys,
            EncryptedKeys,
            WrappedKeys,
            Nonce,
            Ciphertext,
            SenderSignature,
            SenderPubKey,
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
                            "version" => Ok(GeneratedField::Version),
                            "algorithmId" | "algorithm_id" => Ok(GeneratedField::AlgorithmId),
                            "algorithmVersion" | "algorithm_version" => Ok(GeneratedField::AlgorithmVersion),
                            "recipientKeyIds" | "recipient_key_ids" => Ok(GeneratedField::RecipientKeyIds),
                            "recipientPublicKeys" | "recipient_public_keys" => Ok(GeneratedField::RecipientPublicKeys),
                            "encryptedKeys" | "encrypted_keys" => Ok(GeneratedField::EncryptedKeys),
                            "wrappedKeys" | "wrapped_keys" => Ok(GeneratedField::WrappedKeys),
                            "nonce" => Ok(GeneratedField::Nonce),
                            "ciphertext" => Ok(GeneratedField::Ciphertext),
                            "senderSignature" | "sender_signature" => Ok(GeneratedField::SenderSignature),
                            "senderPubKey" | "sender_pub_key" => Ok(GeneratedField::SenderPubKey),
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
            type Value = EncryptedPayloadEnvelope;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.EncryptedPayloadEnvelope")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EncryptedPayloadEnvelope, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut version__ = None;
                let mut algorithm_id__ = None;
                let mut algorithm_version__ = None;
                let mut recipient_key_ids__ = None;
                let mut recipient_public_keys__ = None;
                let mut encrypted_keys__ = None;
                let mut wrapped_keys__ = None;
                let mut nonce__ = None;
                let mut ciphertext__ = None;
                let mut sender_signature__ = None;
                let mut sender_pub_key__ = None;
                let mut metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AlgorithmId => {
                            if algorithm_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmId"));
                            }
                            algorithm_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AlgorithmVersion => {
                            if algorithm_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmVersion"));
                            }
                            algorithm_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RecipientKeyIds => {
                            if recipient_key_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientKeyIds"));
                            }
                            recipient_key_ids__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecipientPublicKeys => {
                            if recipient_public_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientPublicKeys"));
                            }
                            recipient_public_keys__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
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
                        GeneratedField::WrappedKeys => {
                            if wrapped_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("wrappedKeys"));
                            }
                            wrapped_keys__ = Some(map_.next_value()?);
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
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(EncryptedPayloadEnvelope {
                    version: version__.unwrap_or_default(),
                    algorithm_id: algorithm_id__.unwrap_or_default(),
                    algorithm_version: algorithm_version__.unwrap_or_default(),
                    recipient_key_ids: recipient_key_ids__.unwrap_or_default(),
                    recipient_public_keys: recipient_public_keys__.unwrap_or_default(),
                    encrypted_keys: encrypted_keys__.unwrap_or_default(),
                    wrapped_keys: wrapped_keys__.unwrap_or_default(),
                    nonce: nonce__.unwrap_or_default(),
                    ciphertext: ciphertext__.unwrap_or_default(),
                    sender_signature: sender_signature__.unwrap_or_default(),
                    sender_pub_key: sender_pub_key__.unwrap_or_default(),
                    metadata: metadata__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.EncryptedPayloadEnvelope", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventKeyRegistered {
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
        if !self.fingerprint.is_empty() {
            len += 1;
        }
        if !self.algorithm.is_empty() {
            len += 1;
        }
        if !self.label.is_empty() {
            len += 1;
        }
        if self.registered_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.EventKeyRegistered", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.fingerprint.is_empty() {
            struct_ser.serialize_field("fingerprint", &self.fingerprint)?;
        }
        if !self.algorithm.is_empty() {
            struct_ser.serialize_field("algorithm", &self.algorithm)?;
        }
        if !self.label.is_empty() {
            struct_ser.serialize_field("label", &self.label)?;
        }
        if self.registered_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("registeredAt", ToString::to_string(&self.registered_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventKeyRegistered {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "fingerprint",
            "algorithm",
            "label",
            "registered_at",
            "registeredAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Fingerprint,
            Algorithm,
            Label,
            RegisteredAt,
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
                            "fingerprint" => Ok(GeneratedField::Fingerprint),
                            "algorithm" => Ok(GeneratedField::Algorithm),
                            "label" => Ok(GeneratedField::Label),
                            "registeredAt" | "registered_at" => Ok(GeneratedField::RegisteredAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventKeyRegistered;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.EventKeyRegistered")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventKeyRegistered, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut fingerprint__ = None;
                let mut algorithm__ = None;
                let mut label__ = None;
                let mut registered_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Fingerprint => {
                            if fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fingerprint"));
                            }
                            fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Algorithm => {
                            if algorithm__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithm"));
                            }
                            algorithm__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Label => {
                            if label__.is_some() {
                                return Err(serde::de::Error::duplicate_field("label"));
                            }
                            label__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RegisteredAt => {
                            if registered_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registeredAt"));
                            }
                            registered_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventKeyRegistered {
                    address: address__.unwrap_or_default(),
                    fingerprint: fingerprint__.unwrap_or_default(),
                    algorithm: algorithm__.unwrap_or_default(),
                    label: label__.unwrap_or_default(),
                    registered_at: registered_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.EventKeyRegistered", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventKeyRevoked {
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
        if !self.fingerprint.is_empty() {
            len += 1;
        }
        if !self.revoked_by.is_empty() {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.EventKeyRevoked", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.fingerprint.is_empty() {
            struct_ser.serialize_field("fingerprint", &self.fingerprint)?;
        }
        if !self.revoked_by.is_empty() {
            struct_ser.serialize_field("revokedBy", &self.revoked_by)?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventKeyRevoked {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "fingerprint",
            "revoked_by",
            "revokedBy",
            "revoked_at",
            "revokedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Fingerprint,
            RevokedBy,
            RevokedAt,
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
                            "fingerprint" => Ok(GeneratedField::Fingerprint),
                            "revokedBy" | "revoked_by" => Ok(GeneratedField::RevokedBy),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventKeyRevoked;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.EventKeyRevoked")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventKeyRevoked, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut fingerprint__ = None;
                let mut revoked_by__ = None;
                let mut revoked_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Fingerprint => {
                            if fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fingerprint"));
                            }
                            fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedBy => {
                            if revoked_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedBy"));
                            }
                            revoked_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventKeyRevoked {
                    address: address__.unwrap_or_default(),
                    fingerprint: fingerprint__.unwrap_or_default(),
                    revoked_by: revoked_by__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.EventKeyRevoked", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventKeyUpdated {
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
        if !self.fingerprint.is_empty() {
            len += 1;
        }
        if !self.field.is_empty() {
            len += 1;
        }
        if !self.old_value.is_empty() {
            len += 1;
        }
        if !self.new_value.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.EventKeyUpdated", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.fingerprint.is_empty() {
            struct_ser.serialize_field("fingerprint", &self.fingerprint)?;
        }
        if !self.field.is_empty() {
            struct_ser.serialize_field("field", &self.field)?;
        }
        if !self.old_value.is_empty() {
            struct_ser.serialize_field("oldValue", &self.old_value)?;
        }
        if !self.new_value.is_empty() {
            struct_ser.serialize_field("newValue", &self.new_value)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventKeyUpdated {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "fingerprint",
            "field",
            "old_value",
            "oldValue",
            "new_value",
            "newValue",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Fingerprint,
            Field,
            OldValue,
            NewValue,
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
                            "fingerprint" => Ok(GeneratedField::Fingerprint),
                            "field" => Ok(GeneratedField::Field),
                            "oldValue" | "old_value" => Ok(GeneratedField::OldValue),
                            "newValue" | "new_value" => Ok(GeneratedField::NewValue),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventKeyUpdated;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.EventKeyUpdated")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventKeyUpdated, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut fingerprint__ = None;
                let mut field__ = None;
                let mut old_value__ = None;
                let mut new_value__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Fingerprint => {
                            if fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fingerprint"));
                            }
                            fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Field => {
                            if field__.is_some() {
                                return Err(serde::de::Error::duplicate_field("field"));
                            }
                            field__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OldValue => {
                            if old_value__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldValue"));
                            }
                            old_value__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewValue => {
                            if new_value__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newValue"));
                            }
                            new_value__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventKeyUpdated {
                    address: address__.unwrap_or_default(),
                    fingerprint: fingerprint__.unwrap_or_default(),
                    field: field__.unwrap_or_default(),
                    old_value: old_value__.unwrap_or_default(),
                    new_value: new_value__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.EventKeyUpdated", FIELDS, GeneratedVisitor)
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
        if !self.recipient_keys.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.recipient_keys.is_empty() {
            struct_ser.serialize_field("recipientKeys", &self.recipient_keys)?;
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
            "recipient_keys",
            "recipientKeys",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            RecipientKeys,
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
                            "recipientKeys" | "recipient_keys" => Ok(GeneratedField::RecipientKeys),
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
                formatter.write_str("struct virtengine.encryption.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut recipient_keys__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::RecipientKeys => {
                            if recipient_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientKeys"));
                            }
                            recipient_keys__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    recipient_keys: recipient_keys__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterRecipientKey {
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
        if !self.public_key.is_empty() {
            len += 1;
        }
        if !self.algorithm_id.is_empty() {
            len += 1;
        }
        if !self.label.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MsgRegisterRecipientKey", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.public_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("publicKey", pbjson::private::base64::encode(&self.public_key).as_str())?;
        }
        if !self.algorithm_id.is_empty() {
            struct_ser.serialize_field("algorithmId", &self.algorithm_id)?;
        }
        if !self.label.is_empty() {
            struct_ser.serialize_field("label", &self.label)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterRecipientKey {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "public_key",
            "publicKey",
            "algorithm_id",
            "algorithmId",
            "label",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            PublicKey,
            AlgorithmId,
            Label,
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
                            "publicKey" | "public_key" => Ok(GeneratedField::PublicKey),
                            "algorithmId" | "algorithm_id" => Ok(GeneratedField::AlgorithmId),
                            "label" => Ok(GeneratedField::Label),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterRecipientKey;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MsgRegisterRecipientKey")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterRecipientKey, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut public_key__ = None;
                let mut algorithm_id__ = None;
                let mut label__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicKey => {
                            if public_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicKey"));
                            }
                            public_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AlgorithmId => {
                            if algorithm_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmId"));
                            }
                            algorithm_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Label => {
                            if label__.is_some() {
                                return Err(serde::de::Error::duplicate_field("label"));
                            }
                            label__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRegisterRecipientKey {
                    sender: sender__.unwrap_or_default(),
                    public_key: public_key__.unwrap_or_default(),
                    algorithm_id: algorithm_id__.unwrap_or_default(),
                    label: label__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MsgRegisterRecipientKey", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterRecipientKeyResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MsgRegisterRecipientKeyResponse", len)?;
        if !self.key_fingerprint.is_empty() {
            struct_ser.serialize_field("keyFingerprint", &self.key_fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterRecipientKeyResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key_fingerprint",
            "keyFingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            KeyFingerprint,
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
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterRecipientKeyResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MsgRegisterRecipientKeyResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterRecipientKeyResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key_fingerprint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::KeyFingerprint => {
                            if key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyFingerprint"));
                            }
                            key_fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRegisterRecipientKeyResponse {
                    key_fingerprint: key_fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MsgRegisterRecipientKeyResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeRecipientKey {
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
        if !self.key_fingerprint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MsgRevokeRecipientKey", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.key_fingerprint.is_empty() {
            struct_ser.serialize_field("keyFingerprint", &self.key_fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeRecipientKey {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "key_fingerprint",
            "keyFingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            KeyFingerprint,
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
                            "keyFingerprint" | "key_fingerprint" => Ok(GeneratedField::KeyFingerprint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeRecipientKey;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MsgRevokeRecipientKey")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeRecipientKey, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut key_fingerprint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::KeyFingerprint => {
                            if key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyFingerprint"));
                            }
                            key_fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRevokeRecipientKey {
                    sender: sender__.unwrap_or_default(),
                    key_fingerprint: key_fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MsgRevokeRecipientKey", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeRecipientKeyResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MsgRevokeRecipientKeyResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeRecipientKeyResponse {
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
            type Value = MsgRevokeRecipientKeyResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MsgRevokeRecipientKeyResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeRecipientKeyResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRevokeRecipientKeyResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MsgRevokeRecipientKeyResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateKeyLabel {
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
        if !self.key_fingerprint.is_empty() {
            len += 1;
        }
        if !self.label.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MsgUpdateKeyLabel", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.key_fingerprint.is_empty() {
            struct_ser.serialize_field("keyFingerprint", &self.key_fingerprint)?;
        }
        if !self.label.is_empty() {
            struct_ser.serialize_field("label", &self.label)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateKeyLabel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "key_fingerprint",
            "keyFingerprint",
            "label",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            KeyFingerprint,
            Label,
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
                            "keyFingerprint" | "key_fingerprint" => Ok(GeneratedField::KeyFingerprint),
                            "label" => Ok(GeneratedField::Label),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateKeyLabel;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MsgUpdateKeyLabel")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateKeyLabel, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut key_fingerprint__ = None;
                let mut label__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::KeyFingerprint => {
                            if key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyFingerprint"));
                            }
                            key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Label => {
                            if label__.is_some() {
                                return Err(serde::de::Error::duplicate_field("label"));
                            }
                            label__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateKeyLabel {
                    sender: sender__.unwrap_or_default(),
                    key_fingerprint: key_fingerprint__.unwrap_or_default(),
                    label: label__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MsgUpdateKeyLabel", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateKeyLabelResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MsgUpdateKeyLabelResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateKeyLabelResponse {
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
            type Value = MsgUpdateKeyLabelResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MsgUpdateKeyLabelResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateKeyLabelResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateKeyLabelResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MsgUpdateKeyLabelResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MultiRecipientEnvelope {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.version != 0 {
            len += 1;
        }
        if !self.algorithm_id.is_empty() {
            len += 1;
        }
        if self.algorithm_version != 0 {
            len += 1;
        }
        if self.recipient_mode != 0 {
            len += 1;
        }
        if !self.payload_ciphertext.is_empty() {
            len += 1;
        }
        if !self.payload_nonce.is_empty() {
            len += 1;
        }
        if !self.wrapped_keys.is_empty() {
            len += 1;
        }
        if !self.client_signature.is_empty() {
            len += 1;
        }
        if !self.client_id.is_empty() {
            len += 1;
        }
        if !self.user_signature.is_empty() {
            len += 1;
        }
        if !self.user_pub_key.is_empty() {
            len += 1;
        }
        if !self.metadata.is_empty() {
            len += 1;
        }
        if self.committee_epoch != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.MultiRecipientEnvelope", len)?;
        if self.version != 0 {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if !self.algorithm_id.is_empty() {
            struct_ser.serialize_field("algorithmId", &self.algorithm_id)?;
        }
        if self.algorithm_version != 0 {
            struct_ser.serialize_field("algorithmVersion", &self.algorithm_version)?;
        }
        if self.recipient_mode != 0 {
            let v = RecipientMode::try_from(self.recipient_mode)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.recipient_mode)))?;
            struct_ser.serialize_field("recipientMode", &v)?;
        }
        if !self.payload_ciphertext.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("payloadCiphertext", pbjson::private::base64::encode(&self.payload_ciphertext).as_str())?;
        }
        if !self.payload_nonce.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("payloadNonce", pbjson::private::base64::encode(&self.payload_nonce).as_str())?;
        }
        if !self.wrapped_keys.is_empty() {
            struct_ser.serialize_field("wrappedKeys", &self.wrapped_keys)?;
        }
        if !self.client_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("clientSignature", pbjson::private::base64::encode(&self.client_signature).as_str())?;
        }
        if !self.client_id.is_empty() {
            struct_ser.serialize_field("clientId", &self.client_id)?;
        }
        if !self.user_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userSignature", pbjson::private::base64::encode(&self.user_signature).as_str())?;
        }
        if !self.user_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userPubKey", pbjson::private::base64::encode(&self.user_pub_key).as_str())?;
        }
        if !self.metadata.is_empty() {
            struct_ser.serialize_field("metadata", &self.metadata)?;
        }
        if self.committee_epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("committeeEpoch", ToString::to_string(&self.committee_epoch).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MultiRecipientEnvelope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "version",
            "algorithm_id",
            "algorithmId",
            "algorithm_version",
            "algorithmVersion",
            "recipient_mode",
            "recipientMode",
            "payload_ciphertext",
            "payloadCiphertext",
            "payload_nonce",
            "payloadNonce",
            "wrapped_keys",
            "wrappedKeys",
            "client_signature",
            "clientSignature",
            "client_id",
            "clientId",
            "user_signature",
            "userSignature",
            "user_pub_key",
            "userPubKey",
            "metadata",
            "committee_epoch",
            "committeeEpoch",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Version,
            AlgorithmId,
            AlgorithmVersion,
            RecipientMode,
            PayloadCiphertext,
            PayloadNonce,
            WrappedKeys,
            ClientSignature,
            ClientId,
            UserSignature,
            UserPubKey,
            Metadata,
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
                            "version" => Ok(GeneratedField::Version),
                            "algorithmId" | "algorithm_id" => Ok(GeneratedField::AlgorithmId),
                            "algorithmVersion" | "algorithm_version" => Ok(GeneratedField::AlgorithmVersion),
                            "recipientMode" | "recipient_mode" => Ok(GeneratedField::RecipientMode),
                            "payloadCiphertext" | "payload_ciphertext" => Ok(GeneratedField::PayloadCiphertext),
                            "payloadNonce" | "payload_nonce" => Ok(GeneratedField::PayloadNonce),
                            "wrappedKeys" | "wrapped_keys" => Ok(GeneratedField::WrappedKeys),
                            "clientSignature" | "client_signature" => Ok(GeneratedField::ClientSignature),
                            "clientId" | "client_id" => Ok(GeneratedField::ClientId),
                            "userSignature" | "user_signature" => Ok(GeneratedField::UserSignature),
                            "userPubKey" | "user_pub_key" => Ok(GeneratedField::UserPubKey),
                            "metadata" => Ok(GeneratedField::Metadata),
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
            type Value = MultiRecipientEnvelope;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.MultiRecipientEnvelope")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MultiRecipientEnvelope, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut version__ = None;
                let mut algorithm_id__ = None;
                let mut algorithm_version__ = None;
                let mut recipient_mode__ = None;
                let mut payload_ciphertext__ = None;
                let mut payload_nonce__ = None;
                let mut wrapped_keys__ = None;
                let mut client_signature__ = None;
                let mut client_id__ = None;
                let mut user_signature__ = None;
                let mut user_pub_key__ = None;
                let mut metadata__ = None;
                let mut committee_epoch__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AlgorithmId => {
                            if algorithm_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmId"));
                            }
                            algorithm_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AlgorithmVersion => {
                            if algorithm_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmVersion"));
                            }
                            algorithm_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RecipientMode => {
                            if recipient_mode__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientMode"));
                            }
                            recipient_mode__ = Some(map_.next_value::<RecipientMode>()? as i32);
                        }
                        GeneratedField::PayloadCiphertext => {
                            if payload_ciphertext__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payloadCiphertext"));
                            }
                            payload_ciphertext__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PayloadNonce => {
                            if payload_nonce__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payloadNonce"));
                            }
                            payload_nonce__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::WrappedKeys => {
                            if wrapped_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("wrappedKeys"));
                            }
                            wrapped_keys__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClientSignature => {
                            if client_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientSignature"));
                            }
                            client_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ClientId => {
                            if client_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientId"));
                            }
                            client_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UserSignature => {
                            if user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userSignature"));
                            }
                            user_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UserPubKey => {
                            if user_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userPubKey"));
                            }
                            user_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
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
                Ok(MultiRecipientEnvelope {
                    version: version__.unwrap_or_default(),
                    algorithm_id: algorithm_id__.unwrap_or_default(),
                    algorithm_version: algorithm_version__.unwrap_or_default(),
                    recipient_mode: recipient_mode__.unwrap_or_default(),
                    payload_ciphertext: payload_ciphertext__.unwrap_or_default(),
                    payload_nonce: payload_nonce__.unwrap_or_default(),
                    wrapped_keys: wrapped_keys__.unwrap_or_default(),
                    client_signature: client_signature__.unwrap_or_default(),
                    client_id: client_id__.unwrap_or_default(),
                    user_signature: user_signature__.unwrap_or_default(),
                    user_pub_key: user_pub_key__.unwrap_or_default(),
                    metadata: metadata__.unwrap_or_default(),
                    committee_epoch: committee_epoch__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.MultiRecipientEnvelope", FIELDS, GeneratedVisitor)
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
        if self.max_recipients_per_envelope != 0 {
            len += 1;
        }
        if self.max_keys_per_account != 0 {
            len += 1;
        }
        if !self.allowed_algorithms.is_empty() {
            len += 1;
        }
        if self.require_signature {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.Params", len)?;
        if self.max_recipients_per_envelope != 0 {
            struct_ser.serialize_field("maxRecipientsPerEnvelope", &self.max_recipients_per_envelope)?;
        }
        if self.max_keys_per_account != 0 {
            struct_ser.serialize_field("maxKeysPerAccount", &self.max_keys_per_account)?;
        }
        if !self.allowed_algorithms.is_empty() {
            struct_ser.serialize_field("allowedAlgorithms", &self.allowed_algorithms)?;
        }
        if self.require_signature {
            struct_ser.serialize_field("requireSignature", &self.require_signature)?;
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
            "max_recipients_per_envelope",
            "maxRecipientsPerEnvelope",
            "max_keys_per_account",
            "maxKeysPerAccount",
            "allowed_algorithms",
            "allowedAlgorithms",
            "require_signature",
            "requireSignature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MaxRecipientsPerEnvelope,
            MaxKeysPerAccount,
            AllowedAlgorithms,
            RequireSignature,
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
                            "maxRecipientsPerEnvelope" | "max_recipients_per_envelope" => Ok(GeneratedField::MaxRecipientsPerEnvelope),
                            "maxKeysPerAccount" | "max_keys_per_account" => Ok(GeneratedField::MaxKeysPerAccount),
                            "allowedAlgorithms" | "allowed_algorithms" => Ok(GeneratedField::AllowedAlgorithms),
                            "requireSignature" | "require_signature" => Ok(GeneratedField::RequireSignature),
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
                formatter.write_str("struct virtengine.encryption.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut max_recipients_per_envelope__ = None;
                let mut max_keys_per_account__ = None;
                let mut allowed_algorithms__ = None;
                let mut require_signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MaxRecipientsPerEnvelope => {
                            if max_recipients_per_envelope__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRecipientsPerEnvelope"));
                            }
                            max_recipients_per_envelope__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxKeysPerAccount => {
                            if max_keys_per_account__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxKeysPerAccount"));
                            }
                            max_keys_per_account__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AllowedAlgorithms => {
                            if allowed_algorithms__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowedAlgorithms"));
                            }
                            allowed_algorithms__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireSignature => {
                            if require_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireSignature"));
                            }
                            require_signature__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Params {
                    max_recipients_per_envelope: max_recipients_per_envelope__.unwrap_or_default(),
                    max_keys_per_account: max_keys_per_account__.unwrap_or_default(),
                    allowed_algorithms: allowed_algorithms__.unwrap_or_default(),
                    require_signature: require_signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAlgorithmsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryAlgorithmsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAlgorithmsRequest {
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
            type Value = QueryAlgorithmsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryAlgorithmsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAlgorithmsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryAlgorithmsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryAlgorithmsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAlgorithmsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.algorithms.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryAlgorithmsResponse", len)?;
        if !self.algorithms.is_empty() {
            struct_ser.serialize_field("algorithms", &self.algorithms)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAlgorithmsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "algorithms",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Algorithms,
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
                            "algorithms" => Ok(GeneratedField::Algorithms),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAlgorithmsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryAlgorithmsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAlgorithmsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut algorithms__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Algorithms => {
                            if algorithms__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithms"));
                            }
                            algorithms__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAlgorithmsResponse {
                    algorithms: algorithms__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryAlgorithmsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryKeyByFingerprintRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryKeyByFingerprintRequest", len)?;
        if !self.fingerprint.is_empty() {
            struct_ser.serialize_field("fingerprint", &self.fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryKeyByFingerprintRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "fingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Fingerprint,
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
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryKeyByFingerprintRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryKeyByFingerprintRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryKeyByFingerprintRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut fingerprint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Fingerprint => {
                            if fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fingerprint"));
                            }
                            fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryKeyByFingerprintRequest {
                    fingerprint: fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryKeyByFingerprintRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryKeyByFingerprintResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.key.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryKeyByFingerprintResponse", len)?;
        if let Some(v) = self.key.as_ref() {
            struct_ser.serialize_field("key", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryKeyByFingerprintResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Key,
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
                            "key" => Ok(GeneratedField::Key),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryKeyByFingerprintResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryKeyByFingerprintResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryKeyByFingerprintResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryKeyByFingerprintResponse {
                    key: key__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryKeyByFingerprintResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.encryption.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.encryption.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRecipientKeyRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryRecipientKeyRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRecipientKeyRequest {
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
            type Value = QueryRecipientKeyRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryRecipientKeyRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRecipientKeyRequest, V::Error>
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
                Ok(QueryRecipientKeyRequest {
                    address: address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryRecipientKeyRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRecipientKeyResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.keys.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryRecipientKeyResponse", len)?;
        if !self.keys.is_empty() {
            struct_ser.serialize_field("keys", &self.keys)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRecipientKeyResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "keys",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Keys,
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
                            "keys" => Ok(GeneratedField::Keys),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryRecipientKeyResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryRecipientKeyResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRecipientKeyResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut keys__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Keys => {
                            if keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keys"));
                            }
                            keys__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryRecipientKeyResponse {
                    keys: keys__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryRecipientKeyResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidateEnvelopeRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.envelope.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryValidateEnvelopeRequest", len)?;
        if let Some(v) = self.envelope.as_ref() {
            struct_ser.serialize_field("envelope", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidateEnvelopeRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "envelope",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Envelope,
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
                            "envelope" => Ok(GeneratedField::Envelope),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryValidateEnvelopeRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryValidateEnvelopeRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidateEnvelopeRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut envelope__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Envelope => {
                            if envelope__.is_some() {
                                return Err(serde::de::Error::duplicate_field("envelope"));
                            }
                            envelope__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryValidateEnvelopeRequest {
                    envelope: envelope__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryValidateEnvelopeRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidateEnvelopeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.valid {
            len += 1;
        }
        if !self.error.is_empty() {
            len += 1;
        }
        if self.recipient_count != 0 {
            len += 1;
        }
        if !self.algorithm.is_empty() {
            len += 1;
        }
        if self.signature_valid {
            len += 1;
        }
        if self.all_keys_registered {
            len += 1;
        }
        if !self.missing_keys.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.QueryValidateEnvelopeResponse", len)?;
        if self.valid {
            struct_ser.serialize_field("valid", &self.valid)?;
        }
        if !self.error.is_empty() {
            struct_ser.serialize_field("error", &self.error)?;
        }
        if self.recipient_count != 0 {
            struct_ser.serialize_field("recipientCount", &self.recipient_count)?;
        }
        if !self.algorithm.is_empty() {
            struct_ser.serialize_field("algorithm", &self.algorithm)?;
        }
        if self.signature_valid {
            struct_ser.serialize_field("signatureValid", &self.signature_valid)?;
        }
        if self.all_keys_registered {
            struct_ser.serialize_field("allKeysRegistered", &self.all_keys_registered)?;
        }
        if !self.missing_keys.is_empty() {
            struct_ser.serialize_field("missingKeys", &self.missing_keys)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidateEnvelopeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "valid",
            "error",
            "recipient_count",
            "recipientCount",
            "algorithm",
            "signature_valid",
            "signatureValid",
            "all_keys_registered",
            "allKeysRegistered",
            "missing_keys",
            "missingKeys",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Valid,
            Error,
            RecipientCount,
            Algorithm,
            SignatureValid,
            AllKeysRegistered,
            MissingKeys,
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
                            "valid" => Ok(GeneratedField::Valid),
                            "error" => Ok(GeneratedField::Error),
                            "recipientCount" | "recipient_count" => Ok(GeneratedField::RecipientCount),
                            "algorithm" => Ok(GeneratedField::Algorithm),
                            "signatureValid" | "signature_valid" => Ok(GeneratedField::SignatureValid),
                            "allKeysRegistered" | "all_keys_registered" => Ok(GeneratedField::AllKeysRegistered),
                            "missingKeys" | "missing_keys" => Ok(GeneratedField::MissingKeys),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryValidateEnvelopeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.QueryValidateEnvelopeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidateEnvelopeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut valid__ = None;
                let mut error__ = None;
                let mut recipient_count__ = None;
                let mut algorithm__ = None;
                let mut signature_valid__ = None;
                let mut all_keys_registered__ = None;
                let mut missing_keys__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Valid => {
                            if valid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("valid"));
                            }
                            valid__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Error => {
                            if error__.is_some() {
                                return Err(serde::de::Error::duplicate_field("error"));
                            }
                            error__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecipientCount => {
                            if recipient_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientCount"));
                            }
                            recipient_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Algorithm => {
                            if algorithm__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithm"));
                            }
                            algorithm__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SignatureValid => {
                            if signature_valid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signatureValid"));
                            }
                            signature_valid__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllKeysRegistered => {
                            if all_keys_registered__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allKeysRegistered"));
                            }
                            all_keys_registered__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MissingKeys => {
                            if missing_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("missingKeys"));
                            }
                            missing_keys__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryValidateEnvelopeResponse {
                    valid: valid__.unwrap_or_default(),
                    error: error__.unwrap_or_default(),
                    recipient_count: recipient_count__.unwrap_or_default(),
                    algorithm: algorithm__.unwrap_or_default(),
                    signature_valid: signature_valid__.unwrap_or_default(),
                    all_keys_registered: all_keys_registered__.unwrap_or_default(),
                    missing_keys: missing_keys__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.QueryValidateEnvelopeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RecipientKeyRecord {
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
        if !self.public_key.is_empty() {
            len += 1;
        }
        if !self.key_fingerprint.is_empty() {
            len += 1;
        }
        if !self.algorithm_id.is_empty() {
            len += 1;
        }
        if self.registered_at != 0 {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        if !self.label.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.RecipientKeyRecord", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.public_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("publicKey", pbjson::private::base64::encode(&self.public_key).as_str())?;
        }
        if !self.key_fingerprint.is_empty() {
            struct_ser.serialize_field("keyFingerprint", &self.key_fingerprint)?;
        }
        if !self.algorithm_id.is_empty() {
            struct_ser.serialize_field("algorithmId", &self.algorithm_id)?;
        }
        if self.registered_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("registeredAt", ToString::to_string(&self.registered_at).as_str())?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        if !self.label.is_empty() {
            struct_ser.serialize_field("label", &self.label)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RecipientKeyRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "public_key",
            "publicKey",
            "key_fingerprint",
            "keyFingerprint",
            "algorithm_id",
            "algorithmId",
            "registered_at",
            "registeredAt",
            "revoked_at",
            "revokedAt",
            "label",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            PublicKey,
            KeyFingerprint,
            AlgorithmId,
            RegisteredAt,
            RevokedAt,
            Label,
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
                            "publicKey" | "public_key" => Ok(GeneratedField::PublicKey),
                            "keyFingerprint" | "key_fingerprint" => Ok(GeneratedField::KeyFingerprint),
                            "algorithmId" | "algorithm_id" => Ok(GeneratedField::AlgorithmId),
                            "registeredAt" | "registered_at" => Ok(GeneratedField::RegisteredAt),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            "label" => Ok(GeneratedField::Label),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RecipientKeyRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.RecipientKeyRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RecipientKeyRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut public_key__ = None;
                let mut key_fingerprint__ = None;
                let mut algorithm_id__ = None;
                let mut registered_at__ = None;
                let mut revoked_at__ = None;
                let mut label__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicKey => {
                            if public_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicKey"));
                            }
                            public_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::KeyFingerprint => {
                            if key_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keyFingerprint"));
                            }
                            key_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AlgorithmId => {
                            if algorithm_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmId"));
                            }
                            algorithm_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RegisteredAt => {
                            if registered_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registeredAt"));
                            }
                            registered_at__ = 
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
                        GeneratedField::Label => {
                            if label__.is_some() {
                                return Err(serde::de::Error::duplicate_field("label"));
                            }
                            label__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(RecipientKeyRecord {
                    address: address__.unwrap_or_default(),
                    public_key: public_key__.unwrap_or_default(),
                    key_fingerprint: key_fingerprint__.unwrap_or_default(),
                    algorithm_id: algorithm_id__.unwrap_or_default(),
                    registered_at: registered_at__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                    label: label__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.RecipientKeyRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RecipientMode {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "RECIPIENT_MODE_UNSPECIFIED",
            Self::FullValidatorSet => "RECIPIENT_MODE_FULL_VALIDATOR_SET",
            Self::Committee => "RECIPIENT_MODE_COMMITTEE",
            Self::Specific => "RECIPIENT_MODE_SPECIFIC",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for RecipientMode {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "RECIPIENT_MODE_UNSPECIFIED",
            "RECIPIENT_MODE_FULL_VALIDATOR_SET",
            "RECIPIENT_MODE_COMMITTEE",
            "RECIPIENT_MODE_SPECIFIC",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RecipientMode;

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
                    "RECIPIENT_MODE_UNSPECIFIED" => Ok(RecipientMode::Unspecified),
                    "RECIPIENT_MODE_FULL_VALIDATOR_SET" => Ok(RecipientMode::FullValidatorSet),
                    "RECIPIENT_MODE_COMMITTEE" => Ok(RecipientMode::Committee),
                    "RECIPIENT_MODE_SPECIFIC" => Ok(RecipientMode::Specific),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for WrappedKeyEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.recipient_id.is_empty() {
            len += 1;
        }
        if !self.wrapped_key.is_empty() {
            len += 1;
        }
        if !self.algorithm.is_empty() {
            len += 1;
        }
        if !self.ephemeral_pub_key.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.encryption.v1.WrappedKeyEntry", len)?;
        if !self.recipient_id.is_empty() {
            struct_ser.serialize_field("recipientId", &self.recipient_id)?;
        }
        if !self.wrapped_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("wrappedKey", pbjson::private::base64::encode(&self.wrapped_key).as_str())?;
        }
        if !self.algorithm.is_empty() {
            struct_ser.serialize_field("algorithm", &self.algorithm)?;
        }
        if !self.ephemeral_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("ephemeralPubKey", pbjson::private::base64::encode(&self.ephemeral_pub_key).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for WrappedKeyEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "recipient_id",
            "recipientId",
            "wrapped_key",
            "wrappedKey",
            "algorithm",
            "ephemeral_pub_key",
            "ephemeralPubKey",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RecipientId,
            WrappedKey,
            Algorithm,
            EphemeralPubKey,
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
                            "recipientId" | "recipient_id" => Ok(GeneratedField::RecipientId),
                            "wrappedKey" | "wrapped_key" => Ok(GeneratedField::WrappedKey),
                            "algorithm" => Ok(GeneratedField::Algorithm),
                            "ephemeralPubKey" | "ephemeral_pub_key" => Ok(GeneratedField::EphemeralPubKey),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = WrappedKeyEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.encryption.v1.WrappedKeyEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<WrappedKeyEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut recipient_id__ = None;
                let mut wrapped_key__ = None;
                let mut algorithm__ = None;
                let mut ephemeral_pub_key__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RecipientId => {
                            if recipient_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientId"));
                            }
                            recipient_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::WrappedKey => {
                            if wrapped_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("wrappedKey"));
                            }
                            wrapped_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Algorithm => {
                            if algorithm__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithm"));
                            }
                            algorithm__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EphemeralPubKey => {
                            if ephemeral_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ephemeralPubKey"));
                            }
                            ephemeral_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(WrappedKeyEntry {
                    recipient_id: recipient_id__.unwrap_or_default(),
                    wrapped_key: wrapped_key__.unwrap_or_default(),
                    algorithm: algorithm__.unwrap_or_default(),
                    ephemeral_pub_key: ephemeral_pub_key__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.encryption.v1.WrappedKeyEntry", FIELDS, GeneratedVisitor)
    }
}

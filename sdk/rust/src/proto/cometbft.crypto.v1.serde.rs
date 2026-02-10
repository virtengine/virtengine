// @generated
impl serde::Serialize for DominoOp {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.key.is_empty() {
            len += 1;
        }
        if !self.input.is_empty() {
            len += 1;
        }
        if !self.output.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("cometbft.crypto.v1.DominoOp", len)?;
        if !self.key.is_empty() {
            struct_ser.serialize_field("key", &self.key)?;
        }
        if !self.input.is_empty() {
            struct_ser.serialize_field("input", &self.input)?;
        }
        if !self.output.is_empty() {
            struct_ser.serialize_field("output", &self.output)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DominoOp {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key",
            "input",
            "output",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Key,
            Input,
            Output,
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
                            "input" => Ok(GeneratedField::Input),
                            "output" => Ok(GeneratedField::Output),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DominoOp;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct cometbft.crypto.v1.DominoOp")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DominoOp, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key__ = None;
                let mut input__ = None;
                let mut output__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Input => {
                            if input__.is_some() {
                                return Err(serde::de::Error::duplicate_field("input"));
                            }
                            input__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Output => {
                            if output__.is_some() {
                                return Err(serde::de::Error::duplicate_field("output"));
                            }
                            output__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(DominoOp {
                    key: key__.unwrap_or_default(),
                    input: input__.unwrap_or_default(),
                    output: output__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("cometbft.crypto.v1.DominoOp", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Proof {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.total != 0 {
            len += 1;
        }
        if self.index != 0 {
            len += 1;
        }
        if !self.leaf_hash.is_empty() {
            len += 1;
        }
        if !self.aunts.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("cometbft.crypto.v1.Proof", len)?;
        if self.total != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("total", ToString::to_string(&self.total).as_str())?;
        }
        if self.index != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("index", ToString::to_string(&self.index).as_str())?;
        }
        if !self.leaf_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("leafHash", pbjson::private::base64::encode(&self.leaf_hash).as_str())?;
        }
        if !self.aunts.is_empty() {
            struct_ser.serialize_field("aunts", &self.aunts.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Proof {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "total",
            "index",
            "leaf_hash",
            "leafHash",
            "aunts",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Total,
            Index,
            LeafHash,
            Aunts,
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
                            "total" => Ok(GeneratedField::Total),
                            "index" => Ok(GeneratedField::Index),
                            "leafHash" | "leaf_hash" => Ok(GeneratedField::LeafHash),
                            "aunts" => Ok(GeneratedField::Aunts),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Proof;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct cometbft.crypto.v1.Proof")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Proof, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut total__ = None;
                let mut index__ = None;
                let mut leaf_hash__ = None;
                let mut aunts__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Total => {
                            if total__.is_some() {
                                return Err(serde::de::Error::duplicate_field("total"));
                            }
                            total__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Index => {
                            if index__.is_some() {
                                return Err(serde::de::Error::duplicate_field("index"));
                            }
                            index__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LeafHash => {
                            if leaf_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leafHash"));
                            }
                            leaf_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Aunts => {
                            if aunts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("aunts"));
                            }
                            aunts__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                    }
                }
                Ok(Proof {
                    total: total__.unwrap_or_default(),
                    index: index__.unwrap_or_default(),
                    leaf_hash: leaf_hash__.unwrap_or_default(),
                    aunts: aunts__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("cometbft.crypto.v1.Proof", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProofOp {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.r#type.is_empty() {
            len += 1;
        }
        if !self.key.is_empty() {
            len += 1;
        }
        if !self.data.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("cometbft.crypto.v1.ProofOp", len)?;
        if !self.r#type.is_empty() {
            struct_ser.serialize_field("type", &self.r#type)?;
        }
        if !self.key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("key", pbjson::private::base64::encode(&self.key).as_str())?;
        }
        if !self.data.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("data", pbjson::private::base64::encode(&self.data).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProofOp {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "type",
            "key",
            "data",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Type,
            Key,
            Data,
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
                            "type" => Ok(GeneratedField::Type),
                            "key" => Ok(GeneratedField::Key),
                            "data" => Ok(GeneratedField::Data),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProofOp;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct cometbft.crypto.v1.ProofOp")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ProofOp, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut r#type__ = None;
                let mut key__ = None;
                let mut data__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Type => {
                            if r#type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("type"));
                            }
                            r#type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Data => {
                            if data__.is_some() {
                                return Err(serde::de::Error::duplicate_field("data"));
                            }
                            data__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ProofOp {
                    r#type: r#type__.unwrap_or_default(),
                    key: key__.unwrap_or_default(),
                    data: data__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("cometbft.crypto.v1.ProofOp", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProofOps {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.ops.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("cometbft.crypto.v1.ProofOps", len)?;
        if !self.ops.is_empty() {
            struct_ser.serialize_field("ops", &self.ops)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProofOps {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ops",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Ops,
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
                            "ops" => Ok(GeneratedField::Ops),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProofOps;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct cometbft.crypto.v1.ProofOps")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ProofOps, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut ops__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Ops => {
                            if ops__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ops"));
                            }
                            ops__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ProofOps {
                    ops: ops__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("cometbft.crypto.v1.ProofOps", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PublicKey {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.sum.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("cometbft.crypto.v1.PublicKey", len)?;
        if let Some(v) = self.sum.as_ref() {
            match v {
                public_key::Sum::Ed25519(v) => {
                    #[allow(clippy::needless_borrow)]
                    #[allow(clippy::needless_borrows_for_generic_args)]
                    struct_ser.serialize_field("ed25519", pbjson::private::base64::encode(&v).as_str())?;
                }
                public_key::Sum::Secp256k1(v) => {
                    #[allow(clippy::needless_borrow)]
                    #[allow(clippy::needless_borrows_for_generic_args)]
                    struct_ser.serialize_field("secp256k1", pbjson::private::base64::encode(&v).as_str())?;
                }
                public_key::Sum::Bls12381(v) => {
                    #[allow(clippy::needless_borrow)]
                    #[allow(clippy::needless_borrows_for_generic_args)]
                    struct_ser.serialize_field("bls12381", pbjson::private::base64::encode(&v).as_str())?;
                }
            }
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PublicKey {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ed25519",
            "secp256k1",
            "bls12381",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Ed25519,
            Secp256k1,
            Bls12381,
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
                            "ed25519" => Ok(GeneratedField::Ed25519),
                            "secp256k1" => Ok(GeneratedField::Secp256k1),
                            "bls12381" => Ok(GeneratedField::Bls12381),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PublicKey;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct cometbft.crypto.v1.PublicKey")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PublicKey, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sum__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Ed25519 => {
                            if sum__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ed25519"));
                            }
                            sum__ = map_.next_value::<::std::option::Option<::pbjson::private::BytesDeserialize<_>>>()?.map(|x| public_key::Sum::Ed25519(x.0));
                        }
                        GeneratedField::Secp256k1 => {
                            if sum__.is_some() {
                                return Err(serde::de::Error::duplicate_field("secp256k1"));
                            }
                            sum__ = map_.next_value::<::std::option::Option<::pbjson::private::BytesDeserialize<_>>>()?.map(|x| public_key::Sum::Secp256k1(x.0));
                        }
                        GeneratedField::Bls12381 => {
                            if sum__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bls12381"));
                            }
                            sum__ = map_.next_value::<::std::option::Option<::pbjson::private::BytesDeserialize<_>>>()?.map(|x| public_key::Sum::Bls12381(x.0));
                        }
                    }
                }
                Ok(PublicKey {
                    sum: sum__,
                })
            }
        }
        deserializer.deserialize_struct("cometbft.crypto.v1.PublicKey", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValueOp {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.key.is_empty() {
            len += 1;
        }
        if self.proof.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("cometbft.crypto.v1.ValueOp", len)?;
        if !self.key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("key", pbjson::private::base64::encode(&self.key).as_str())?;
        }
        if let Some(v) = self.proof.as_ref() {
            struct_ser.serialize_field("proof", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValueOp {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key",
            "proof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Key,
            Proof,
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
                            "proof" => Ok(GeneratedField::Proof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValueOp;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct cometbft.crypto.v1.ValueOp")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValueOp, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key__ = None;
                let mut proof__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Proof => {
                            if proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proof"));
                            }
                            proof__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ValueOp {
                    key: key__.unwrap_or_default(),
                    proof: proof__,
                })
            }
        }
        deserializer.deserialize_struct("cometbft.crypto.v1.ValueOp", FIELDS, GeneratedVisitor)
    }
}

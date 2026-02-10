// @generated
impl serde::Serialize for Account {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.scope != 0 {
            len += 1;
        }
        if !self.xid.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.id.v1.Account", len)?;
        if self.scope != 0 {
            let v = Scope::try_from(self.scope)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope)))?;
            struct_ser.serialize_field("scope", &v)?;
        }
        if !self.xid.is_empty() {
            struct_ser.serialize_field("xid", &self.xid)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Account {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope",
            "xid",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Scope,
            Xid,
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
                            "scope" => Ok(GeneratedField::Scope),
                            "xid" => Ok(GeneratedField::Xid),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Account;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.id.v1.Account")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Account, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope__ = None;
                let mut xid__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Scope => {
                            if scope__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scope"));
                            }
                            scope__ = Some(map_.next_value::<Scope>()? as i32);
                        }
                        GeneratedField::Xid => {
                            if xid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("xid"));
                            }
                            xid__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Account {
                    scope: scope__.unwrap_or_default(),
                    xid: xid__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.id.v1.Account", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Payment {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.aid.is_some() {
            len += 1;
        }
        if !self.xid.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.id.v1.Payment", len)?;
        if let Some(v) = self.aid.as_ref() {
            struct_ser.serialize_field("aid", v)?;
        }
        if !self.xid.is_empty() {
            struct_ser.serialize_field("xid", &self.xid)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Payment {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "aid",
            "xid",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Aid,
            Xid,
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
                            "aid" => Ok(GeneratedField::Aid),
                            "xid" => Ok(GeneratedField::Xid),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Payment;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.id.v1.Payment")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Payment, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut aid__ = None;
                let mut xid__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Aid => {
                            if aid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("aid"));
                            }
                            aid__ = map_.next_value()?;
                        }
                        GeneratedField::Xid => {
                            if xid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("xid"));
                            }
                            xid__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Payment {
                    aid: aid__,
                    xid: xid__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.id.v1.Payment", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Scope {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Invalid => "invalid",
            Self::Deployment => "deployment",
            Self::Bid => "bid",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for Scope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "invalid",
            "deployment",
            "bid",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Scope;

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
                    "invalid" => Ok(Scope::Invalid),
                    "deployment" => Ok(Scope::Deployment),
                    "bid" => Ok(Scope::Bid),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}

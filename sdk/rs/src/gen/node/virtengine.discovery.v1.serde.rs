// @generated
impl serde::Serialize for Akash {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.client_info.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.discovery.v1.Akash", len)?;
        if let Some(v) = self.client_info.as_ref() {
            struct_ser.serialize_field("clientInfo", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Akash {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "client_info",
            "clientInfo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = Akash;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.discovery.v1.Akash")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Akash, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut client_info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClientInfo => {
                            if client_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientInfo"));
                            }
                            client_info__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Akash {
                    client_info: client_info__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.discovery.v1.Akash", FIELDS, GeneratedVisitor)
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
        if !self.api_version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.discovery.v1.ClientInfo", len)?;
        if !self.api_version.is_empty() {
            struct_ser.serialize_field("apiVersion", &self.api_version)?;
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
            "api_version",
            "apiVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ApiVersion,
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
                            "apiVersion" | "api_version" => Ok(GeneratedField::ApiVersion),
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
                formatter.write_str("struct virtengine.discovery.v1.ClientInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ClientInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut api_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ApiVersion => {
                            if api_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("apiVersion"));
                            }
                            api_version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ClientInfo {
                    api_version: api_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.discovery.v1.ClientInfo", FIELDS, GeneratedVisitor)
    }
}

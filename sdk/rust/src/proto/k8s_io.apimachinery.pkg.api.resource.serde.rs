// @generated
impl serde::Serialize for Quantity {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.string.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("k8s_io.apimachinery.pkg.api.resource.Quantity", len)?;
        if let Some(v) = self.string.as_ref() {
            struct_ser.serialize_field("string", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Quantity {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "string",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            String,
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
                            "string" => Ok(GeneratedField::String),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Quantity;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct k8s_io.apimachinery.pkg.api.resource.Quantity")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Quantity, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut string__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::String => {
                            if string__.is_some() {
                                return Err(serde::de::Error::duplicate_field("string"));
                            }
                            string__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Quantity {
                    string: string__,
                })
            }
        }
        deserializer.deserialize_struct("k8s_io.apimachinery.pkg.api.resource.Quantity", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuantityValue {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.string.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("k8s_io.apimachinery.pkg.api.resource.QuantityValue", len)?;
        if let Some(v) = self.string.as_ref() {
            struct_ser.serialize_field("string", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuantityValue {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "string",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            String,
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
                            "string" => Ok(GeneratedField::String),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuantityValue;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct k8s_io.apimachinery.pkg.api.resource.QuantityValue")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuantityValue, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut string__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::String => {
                            if string__.is_some() {
                                return Err(serde::de::Error::duplicate_field("string"));
                            }
                            string__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QuantityValue {
                    string: string__,
                })
            }
        }
        deserializer.deserialize_struct("k8s_io.apimachinery.pkg.api.resource.QuantityValue", FIELDS, GeneratedVisitor)
    }
}

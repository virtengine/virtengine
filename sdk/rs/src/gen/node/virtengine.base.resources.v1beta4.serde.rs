// @generated
impl serde::Serialize for Cpu {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.units.is_some() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.CPU", len)?;
        if let Some(v) = self.units.as_ref() {
            struct_ser.serialize_field("units", v)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Cpu {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "units",
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Units,
            Attributes,
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
                            "units" => Ok(GeneratedField::Units),
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Cpu;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.CPU")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Cpu, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut units__ = None;
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Units => {
                            if units__.is_some() {
                                return Err(serde::de::Error::duplicate_field("units"));
                            }
                            units__ = map_.next_value()?;
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Cpu {
                    units: units__,
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.CPU", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Endpoint {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.kind != 0 {
            len += 1;
        }
        if self.sequence_number != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.Endpoint", len)?;
        if self.kind != 0 {
            let v = endpoint::Kind::try_from(self.kind)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.kind)))?;
            struct_ser.serialize_field("kind", &v)?;
        }
        if self.sequence_number != 0 {
            struct_ser.serialize_field("sequenceNumber", &self.sequence_number)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Endpoint {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "kind",
            "sequence_number",
            "sequenceNumber",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Kind,
            SequenceNumber,
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
                            "kind" => Ok(GeneratedField::Kind),
                            "sequenceNumber" | "sequence_number" => Ok(GeneratedField::SequenceNumber),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Endpoint;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.Endpoint")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Endpoint, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut kind__ = None;
                let mut sequence_number__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Kind => {
                            if kind__.is_some() {
                                return Err(serde::de::Error::duplicate_field("kind"));
                            }
                            kind__ = Some(map_.next_value::<endpoint::Kind>()? as i32);
                        }
                        GeneratedField::SequenceNumber => {
                            if sequence_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sequenceNumber"));
                            }
                            sequence_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Endpoint {
                    kind: kind__.unwrap_or_default(),
                    sequence_number: sequence_number__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.Endpoint", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for endpoint::Kind {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::SharedHttp => "SHARED_HTTP",
            Self::RandomPort => "RANDOM_PORT",
            Self::LeasedIp => "LEASED_IP",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for endpoint::Kind {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "SHARED_HTTP",
            "RANDOM_PORT",
            "LEASED_IP",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = endpoint::Kind;

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
                    "SHARED_HTTP" => Ok(endpoint::Kind::SharedHttp),
                    "RANDOM_PORT" => Ok(endpoint::Kind::RandomPort),
                    "LEASED_IP" => Ok(endpoint::Kind::LeasedIp),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for Gpu {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.units.is_some() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.GPU", len)?;
        if let Some(v) = self.units.as_ref() {
            struct_ser.serialize_field("units", v)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Gpu {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "units",
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Units,
            Attributes,
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
                            "units" => Ok(GeneratedField::Units),
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Gpu;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.GPU")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Gpu, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut units__ = None;
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Units => {
                            if units__.is_some() {
                                return Err(serde::de::Error::duplicate_field("units"));
                            }
                            units__ = map_.next_value()?;
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Gpu {
                    units: units__,
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.GPU", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Memory {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.quantity.is_some() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.Memory", len)?;
        if let Some(v) = self.quantity.as_ref() {
            struct_ser.serialize_field("quantity", v)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Memory {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "quantity",
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Quantity,
            Attributes,
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
                            "quantity" => Ok(GeneratedField::Quantity),
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Memory;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.Memory")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Memory, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut quantity__ = None;
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = map_.next_value()?;
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Memory {
                    quantity: quantity__,
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.Memory", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ResourceValue {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.val.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.ResourceValue", len)?;
        if !self.val.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("val", pbjson::private::base64::encode(&self.val).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ResourceValue {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "val",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Val,
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
                            "val" => Ok(GeneratedField::Val),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ResourceValue;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.ResourceValue")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ResourceValue, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut val__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Val => {
                            if val__.is_some() {
                                return Err(serde::de::Error::duplicate_field("val"));
                            }
                            val__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ResourceValue {
                    val: val__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.ResourceValue", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Resources {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.id != 0 {
            len += 1;
        }
        if self.cpu.is_some() {
            len += 1;
        }
        if self.memory.is_some() {
            len += 1;
        }
        if !self.storage.is_empty() {
            len += 1;
        }
        if self.gpu.is_some() {
            len += 1;
        }
        if !self.endpoints.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.Resources", len)?;
        if self.id != 0 {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if let Some(v) = self.cpu.as_ref() {
            struct_ser.serialize_field("cpu", v)?;
        }
        if let Some(v) = self.memory.as_ref() {
            struct_ser.serialize_field("memory", v)?;
        }
        if !self.storage.is_empty() {
            struct_ser.serialize_field("storage", &self.storage)?;
        }
        if let Some(v) = self.gpu.as_ref() {
            struct_ser.serialize_field("gpu", v)?;
        }
        if !self.endpoints.is_empty() {
            struct_ser.serialize_field("endpoints", &self.endpoints)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Resources {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "cpu",
            "memory",
            "storage",
            "gpu",
            "endpoints",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Cpu,
            Memory,
            Storage,
            Gpu,
            Endpoints,
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
                            "cpu" => Ok(GeneratedField::Cpu),
                            "memory" => Ok(GeneratedField::Memory),
                            "storage" => Ok(GeneratedField::Storage),
                            "gpu" => Ok(GeneratedField::Gpu),
                            "endpoints" => Ok(GeneratedField::Endpoints),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Resources;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.Resources")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Resources, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut cpu__ = None;
                let mut memory__ = None;
                let mut storage__ = None;
                let mut gpu__ = None;
                let mut endpoints__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Cpu => {
                            if cpu__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cpu"));
                            }
                            cpu__ = map_.next_value()?;
                        }
                        GeneratedField::Memory => {
                            if memory__.is_some() {
                                return Err(serde::de::Error::duplicate_field("memory"));
                            }
                            memory__ = map_.next_value()?;
                        }
                        GeneratedField::Storage => {
                            if storage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storage"));
                            }
                            storage__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Gpu => {
                            if gpu__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpu"));
                            }
                            gpu__ = map_.next_value()?;
                        }
                        GeneratedField::Endpoints => {
                            if endpoints__.is_some() {
                                return Err(serde::de::Error::duplicate_field("endpoints"));
                            }
                            endpoints__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Resources {
                    id: id__.unwrap_or_default(),
                    cpu: cpu__,
                    memory: memory__,
                    storage: storage__.unwrap_or_default(),
                    gpu: gpu__,
                    endpoints: endpoints__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.Resources", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Storage {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.name.is_empty() {
            len += 1;
        }
        if self.quantity.is_some() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.base.resources.v1beta4.Storage", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if let Some(v) = self.quantity.as_ref() {
            struct_ser.serialize_field("quantity", v)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Storage {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "quantity",
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Quantity,
            Attributes,
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
                            "name" => Ok(GeneratedField::Name),
                            "quantity" => Ok(GeneratedField::Quantity),
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Storage;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.base.resources.v1beta4.Storage")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Storage, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut quantity__ = None;
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = map_.next_value()?;
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Storage {
                    name: name__.unwrap_or_default(),
                    quantity: quantity__,
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.base.resources.v1beta4.Storage", FIELDS, GeneratedVisitor)
    }
}

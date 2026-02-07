// @generated
impl serde::Serialize for Cpu {
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
        if !self.info.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.CPU", len)?;
        if let Some(v) = self.quantity.as_ref() {
            struct_ser.serialize_field("quantity", v)?;
        }
        if !self.info.is_empty() {
            struct_ser.serialize_field("info", &self.info)?;
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
            "quantity",
            "info",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Quantity,
            Info,
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
                            "info" => Ok(GeneratedField::Info),
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
                formatter.write_str("struct virtengine.inventory.v1.CPU")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Cpu, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut quantity__ = None;
                let mut info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = map_.next_value()?;
                        }
                        GeneratedField::Info => {
                            if info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("info"));
                            }
                            info__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Cpu {
                    quantity: quantity__,
                    info: info__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.CPU", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for CpuInfo {
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
        if !self.vendor.is_empty() {
            len += 1;
        }
        if !self.model.is_empty() {
            len += 1;
        }
        if self.vcores != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.CPUInfo", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if !self.vendor.is_empty() {
            struct_ser.serialize_field("vendor", &self.vendor)?;
        }
        if !self.model.is_empty() {
            struct_ser.serialize_field("model", &self.model)?;
        }
        if self.vcores != 0 {
            struct_ser.serialize_field("vcores", &self.vcores)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for CpuInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "vendor",
            "model",
            "vcores",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Vendor,
            Model,
            Vcores,
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
                            "vendor" => Ok(GeneratedField::Vendor),
                            "model" => Ok(GeneratedField::Model),
                            "vcores" => Ok(GeneratedField::Vcores),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = CpuInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.CPUInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<CpuInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut vendor__ = None;
                let mut model__ = None;
                let mut vcores__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Vendor => {
                            if vendor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vendor"));
                            }
                            vendor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Model => {
                            if model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("model"));
                            }
                            model__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Vcores => {
                            if vcores__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vcores"));
                            }
                            vcores__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(CpuInfo {
                    id: id__.unwrap_or_default(),
                    vendor: vendor__.unwrap_or_default(),
                    model: model__.unwrap_or_default(),
                    vcores: vcores__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.CPUInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Cluster {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.nodes.is_empty() {
            len += 1;
        }
        if !self.storage.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.Cluster", len)?;
        if !self.nodes.is_empty() {
            struct_ser.serialize_field("nodes", &self.nodes)?;
        }
        if !self.storage.is_empty() {
            struct_ser.serialize_field("storage", &self.storage)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Cluster {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "nodes",
            "storage",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Nodes,
            Storage,
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
                            "nodes" => Ok(GeneratedField::Nodes),
                            "storage" => Ok(GeneratedField::Storage),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Cluster;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.Cluster")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Cluster, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut nodes__ = None;
                let mut storage__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Nodes => {
                            if nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodes"));
                            }
                            nodes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Storage => {
                            if storage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storage"));
                            }
                            storage__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Cluster {
                    nodes: nodes__.unwrap_or_default(),
                    storage: storage__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.Cluster", FIELDS, GeneratedVisitor)
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
        if self.quantity.is_some() {
            len += 1;
        }
        if !self.info.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.GPU", len)?;
        if let Some(v) = self.quantity.as_ref() {
            struct_ser.serialize_field("quantity", v)?;
        }
        if !self.info.is_empty() {
            struct_ser.serialize_field("info", &self.info)?;
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
            "quantity",
            "info",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Quantity,
            Info,
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
                            "info" => Ok(GeneratedField::Info),
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
                formatter.write_str("struct virtengine.inventory.v1.GPU")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Gpu, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut quantity__ = None;
                let mut info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = map_.next_value()?;
                        }
                        GeneratedField::Info => {
                            if info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("info"));
                            }
                            info__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Gpu {
                    quantity: quantity__,
                    info: info__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.GPU", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for GpuInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.vendor.is_empty() {
            len += 1;
        }
        if !self.vendor_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.modelid.is_empty() {
            len += 1;
        }
        if !self.interface.is_empty() {
            len += 1;
        }
        if !self.memory_size.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.GPUInfo", len)?;
        if !self.vendor.is_empty() {
            struct_ser.serialize_field("vendor", &self.vendor)?;
        }
        if !self.vendor_id.is_empty() {
            struct_ser.serialize_field("vendorId", &self.vendor_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.modelid.is_empty() {
            struct_ser.serialize_field("modelid", &self.modelid)?;
        }
        if !self.interface.is_empty() {
            struct_ser.serialize_field("interface", &self.interface)?;
        }
        if !self.memory_size.is_empty() {
            struct_ser.serialize_field("memorySize", &self.memory_size)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GpuInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "vendor",
            "vendor_id",
            "vendorId",
            "name",
            "modelid",
            "interface",
            "memory_size",
            "memorySize",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Vendor,
            VendorId,
            Name,
            Modelid,
            Interface,
            MemorySize,
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
                            "vendor" => Ok(GeneratedField::Vendor),
                            "vendorId" | "vendor_id" => Ok(GeneratedField::VendorId),
                            "name" => Ok(GeneratedField::Name),
                            "modelid" => Ok(GeneratedField::Modelid),
                            "interface" => Ok(GeneratedField::Interface),
                            "memorySize" | "memory_size" => Ok(GeneratedField::MemorySize),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GpuInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.GPUInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GpuInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut vendor__ = None;
                let mut vendor_id__ = None;
                let mut name__ = None;
                let mut modelid__ = None;
                let mut interface__ = None;
                let mut memory_size__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Vendor => {
                            if vendor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vendor"));
                            }
                            vendor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::VendorId => {
                            if vendor_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vendorId"));
                            }
                            vendor_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Modelid => {
                            if modelid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelid"));
                            }
                            modelid__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Interface => {
                            if interface__.is_some() {
                                return Err(serde::de::Error::duplicate_field("interface"));
                            }
                            interface__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MemorySize => {
                            if memory_size__.is_some() {
                                return Err(serde::de::Error::duplicate_field("memorySize"));
                            }
                            memory_size__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GpuInfo {
                    vendor: vendor__.unwrap_or_default(),
                    vendor_id: vendor_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    modelid: modelid__.unwrap_or_default(),
                    interface: interface__.unwrap_or_default(),
                    memory_size: memory_size__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.GPUInfo", FIELDS, GeneratedVisitor)
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
        if !self.info.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.Memory", len)?;
        if let Some(v) = self.quantity.as_ref() {
            struct_ser.serialize_field("quantity", v)?;
        }
        if !self.info.is_empty() {
            struct_ser.serialize_field("info", &self.info)?;
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
            "info",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Quantity,
            Info,
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
                            "info" => Ok(GeneratedField::Info),
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
                formatter.write_str("struct virtengine.inventory.v1.Memory")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Memory, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut quantity__ = None;
                let mut info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = map_.next_value()?;
                        }
                        GeneratedField::Info => {
                            if info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("info"));
                            }
                            info__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Memory {
                    quantity: quantity__,
                    info: info__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.Memory", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MemoryInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.vendor.is_empty() {
            len += 1;
        }
        if !self.r#type.is_empty() {
            len += 1;
        }
        if !self.total_size.is_empty() {
            len += 1;
        }
        if !self.speed.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.MemoryInfo", len)?;
        if !self.vendor.is_empty() {
            struct_ser.serialize_field("vendor", &self.vendor)?;
        }
        if !self.r#type.is_empty() {
            struct_ser.serialize_field("type", &self.r#type)?;
        }
        if !self.total_size.is_empty() {
            struct_ser.serialize_field("totalSize", &self.total_size)?;
        }
        if !self.speed.is_empty() {
            struct_ser.serialize_field("speed", &self.speed)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MemoryInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "vendor",
            "type",
            "total_size",
            "totalSize",
            "speed",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Vendor,
            Type,
            TotalSize,
            Speed,
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
                            "vendor" => Ok(GeneratedField::Vendor),
                            "type" => Ok(GeneratedField::Type),
                            "totalSize" | "total_size" => Ok(GeneratedField::TotalSize),
                            "speed" => Ok(GeneratedField::Speed),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MemoryInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.MemoryInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MemoryInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut vendor__ = None;
                let mut r#type__ = None;
                let mut total_size__ = None;
                let mut speed__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Vendor => {
                            if vendor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vendor"));
                            }
                            vendor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Type => {
                            if r#type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("type"));
                            }
                            r#type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalSize => {
                            if total_size__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalSize"));
                            }
                            total_size__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Speed => {
                            if speed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("speed"));
                            }
                            speed__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MemoryInfo {
                    vendor: vendor__.unwrap_or_default(),
                    r#type: r#type__.unwrap_or_default(),
                    total_size: total_size__.unwrap_or_default(),
                    speed: speed__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.MemoryInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Node {
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
        if self.resources.is_some() {
            len += 1;
        }
        if self.capabilities.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.Node", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        if let Some(v) = self.capabilities.as_ref() {
            struct_ser.serialize_field("capabilities", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Node {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "resources",
            "capabilities",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Resources,
            Capabilities,
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
                            "resources" => Ok(GeneratedField::Resources),
                            "capabilities" => Ok(GeneratedField::Capabilities),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Node;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.Node")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Node, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut resources__ = None;
                let mut capabilities__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                        GeneratedField::Capabilities => {
                            if capabilities__.is_some() {
                                return Err(serde::de::Error::duplicate_field("capabilities"));
                            }
                            capabilities__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Node {
                    name: name__.unwrap_or_default(),
                    resources: resources__,
                    capabilities: capabilities__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.Node", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for NodeCapabilities {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.storage_classes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.NodeCapabilities", len)?;
        if !self.storage_classes.is_empty() {
            struct_ser.serialize_field("storageClasses", &self.storage_classes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for NodeCapabilities {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "storage_classes",
            "storageClasses",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            StorageClasses,
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
                            "storageClasses" | "storage_classes" => Ok(GeneratedField::StorageClasses),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = NodeCapabilities;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.NodeCapabilities")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<NodeCapabilities, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut storage_classes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::StorageClasses => {
                            if storage_classes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storageClasses"));
                            }
                            storage_classes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(NodeCapabilities {
                    storage_classes: storage_classes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.NodeCapabilities", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for NodeResources {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.cpu.is_some() {
            len += 1;
        }
        if self.memory.is_some() {
            len += 1;
        }
        if self.gpu.is_some() {
            len += 1;
        }
        if self.ephemeral_storage.is_some() {
            len += 1;
        }
        if self.volumes_attached.is_some() {
            len += 1;
        }
        if self.volumes_mounted.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.NodeResources", len)?;
        if let Some(v) = self.cpu.as_ref() {
            struct_ser.serialize_field("cpu", v)?;
        }
        if let Some(v) = self.memory.as_ref() {
            struct_ser.serialize_field("memory", v)?;
        }
        if let Some(v) = self.gpu.as_ref() {
            struct_ser.serialize_field("gpu", v)?;
        }
        if let Some(v) = self.ephemeral_storage.as_ref() {
            struct_ser.serialize_field("ephemeralStorage", v)?;
        }
        if let Some(v) = self.volumes_attached.as_ref() {
            struct_ser.serialize_field("volumesAttached", v)?;
        }
        if let Some(v) = self.volumes_mounted.as_ref() {
            struct_ser.serialize_field("volumesMounted", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for NodeResources {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cpu",
            "memory",
            "gpu",
            "ephemeral_storage",
            "ephemeralStorage",
            "volumes_attached",
            "volumesAttached",
            "volumes_mounted",
            "volumesMounted",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Cpu,
            Memory,
            Gpu,
            EphemeralStorage,
            VolumesAttached,
            VolumesMounted,
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
                            "cpu" => Ok(GeneratedField::Cpu),
                            "memory" => Ok(GeneratedField::Memory),
                            "gpu" => Ok(GeneratedField::Gpu),
                            "ephemeralStorage" | "ephemeral_storage" => Ok(GeneratedField::EphemeralStorage),
                            "volumesAttached" | "volumes_attached" => Ok(GeneratedField::VolumesAttached),
                            "volumesMounted" | "volumes_mounted" => Ok(GeneratedField::VolumesMounted),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = NodeResources;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.NodeResources")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<NodeResources, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cpu__ = None;
                let mut memory__ = None;
                let mut gpu__ = None;
                let mut ephemeral_storage__ = None;
                let mut volumes_attached__ = None;
                let mut volumes_mounted__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
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
                        GeneratedField::Gpu => {
                            if gpu__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpu"));
                            }
                            gpu__ = map_.next_value()?;
                        }
                        GeneratedField::EphemeralStorage => {
                            if ephemeral_storage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ephemeralStorage"));
                            }
                            ephemeral_storage__ = map_.next_value()?;
                        }
                        GeneratedField::VolumesAttached => {
                            if volumes_attached__.is_some() {
                                return Err(serde::de::Error::duplicate_field("volumesAttached"));
                            }
                            volumes_attached__ = map_.next_value()?;
                        }
                        GeneratedField::VolumesMounted => {
                            if volumes_mounted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("volumesMounted"));
                            }
                            volumes_mounted__ = map_.next_value()?;
                        }
                    }
                }
                Ok(NodeResources {
                    cpu: cpu__,
                    memory: memory__,
                    gpu: gpu__,
                    ephemeral_storage: ephemeral_storage__,
                    volumes_attached: volumes_attached__,
                    volumes_mounted: volumes_mounted__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.NodeResources", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ResourcePair {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.allocatable.is_some() {
            len += 1;
        }
        if self.allocated.is_some() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        if self.capacity.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.ResourcePair", len)?;
        if let Some(v) = self.allocatable.as_ref() {
            struct_ser.serialize_field("allocatable", v)?;
        }
        if let Some(v) = self.allocated.as_ref() {
            struct_ser.serialize_field("allocated", v)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        if let Some(v) = self.capacity.as_ref() {
            struct_ser.serialize_field("capacity", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ResourcePair {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "allocatable",
            "allocated",
            "attributes",
            "capacity",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Allocatable,
            Allocated,
            Attributes,
            Capacity,
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
                            "allocatable" => Ok(GeneratedField::Allocatable),
                            "allocated" => Ok(GeneratedField::Allocated),
                            "attributes" => Ok(GeneratedField::Attributes),
                            "capacity" => Ok(GeneratedField::Capacity),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ResourcePair;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.ResourcePair")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ResourcePair, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut allocatable__ = None;
                let mut allocated__ = None;
                let mut attributes__ = None;
                let mut capacity__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Allocatable => {
                            if allocatable__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocatable"));
                            }
                            allocatable__ = map_.next_value()?;
                        }
                        GeneratedField::Allocated => {
                            if allocated__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocated"));
                            }
                            allocated__ = map_.next_value()?;
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Capacity => {
                            if capacity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("capacity"));
                            }
                            capacity__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ResourcePair {
                    allocatable: allocatable__,
                    allocated: allocated__,
                    attributes: attributes__.unwrap_or_default(),
                    capacity: capacity__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.ResourcePair", FIELDS, GeneratedVisitor)
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
        if self.quantity.is_some() {
            len += 1;
        }
        if self.info.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.Storage", len)?;
        if let Some(v) = self.quantity.as_ref() {
            struct_ser.serialize_field("quantity", v)?;
        }
        if let Some(v) = self.info.as_ref() {
            struct_ser.serialize_field("info", v)?;
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
            "quantity",
            "info",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Quantity,
            Info,
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
                            "info" => Ok(GeneratedField::Info),
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
                formatter.write_str("struct virtengine.inventory.v1.Storage")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Storage, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut quantity__ = None;
                let mut info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = map_.next_value()?;
                        }
                        GeneratedField::Info => {
                            if info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("info"));
                            }
                            info__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Storage {
                    quantity: quantity__,
                    info: info__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.Storage", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for StorageInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.class.is_empty() {
            len += 1;
        }
        if !self.iops.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.inventory.v1.StorageInfo", len)?;
        if !self.class.is_empty() {
            struct_ser.serialize_field("class", &self.class)?;
        }
        if !self.iops.is_empty() {
            struct_ser.serialize_field("iops", &self.iops)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for StorageInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "class",
            "iops",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Class,
            Iops,
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
                            "class" => Ok(GeneratedField::Class),
                            "iops" => Ok(GeneratedField::Iops),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = StorageInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.inventory.v1.StorageInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<StorageInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut class__ = None;
                let mut iops__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Class => {
                            if class__.is_some() {
                                return Err(serde::de::Error::duplicate_field("class"));
                            }
                            class__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Iops => {
                            if iops__.is_some() {
                                return Err(serde::de::Error::duplicate_field("iops"));
                            }
                            iops__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(StorageInfo {
                    class: class__.unwrap_or_default(),
                    iops: iops__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.inventory.v1.StorageInfo", FIELDS, GeneratedVisitor)
    }
}

// @generated
impl serde::Serialize for BidEngineStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.orders != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.BidEngineStatus", len)?;
        if self.orders != 0 {
            struct_ser.serialize_field("orders", &self.orders)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BidEngineStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "orders",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Orders,
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
                            "orders" => Ok(GeneratedField::Orders),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BidEngineStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.BidEngineStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<BidEngineStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut orders__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Orders => {
                            if orders__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orders"));
                            }
                            orders__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(BidEngineStatus {
                    orders: orders__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.BidEngineStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ClusterStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.leases.is_some() {
            len += 1;
        }
        if self.inventory.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.ClusterStatus", len)?;
        if let Some(v) = self.leases.as_ref() {
            struct_ser.serialize_field("leases", v)?;
        }
        if let Some(v) = self.inventory.as_ref() {
            struct_ser.serialize_field("inventory", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ClusterStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "leases",
            "inventory",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Leases,
            Inventory,
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
                            "leases" => Ok(GeneratedField::Leases),
                            "inventory" => Ok(GeneratedField::Inventory),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ClusterStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.ClusterStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ClusterStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut leases__ = None;
                let mut inventory__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Leases => {
                            if leases__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leases"));
                            }
                            leases__ = map_.next_value()?;
                        }
                        GeneratedField::Inventory => {
                            if inventory__.is_some() {
                                return Err(serde::de::Error::duplicate_field("inventory"));
                            }
                            inventory__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ClusterStatus {
                    leases: leases__,
                    inventory: inventory__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.ClusterStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Inventory {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.cluster.is_some() {
            len += 1;
        }
        if self.reservations.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.Inventory", len)?;
        if let Some(v) = self.cluster.as_ref() {
            struct_ser.serialize_field("cluster", v)?;
        }
        if let Some(v) = self.reservations.as_ref() {
            struct_ser.serialize_field("reservations", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Inventory {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster",
            "reservations",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Cluster,
            Reservations,
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
                            "cluster" => Ok(GeneratedField::Cluster),
                            "reservations" => Ok(GeneratedField::Reservations),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Inventory;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.Inventory")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Inventory, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster__ = None;
                let mut reservations__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Cluster => {
                            if cluster__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cluster"));
                            }
                            cluster__ = map_.next_value()?;
                        }
                        GeneratedField::Reservations => {
                            if reservations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reservations"));
                            }
                            reservations__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Inventory {
                    cluster: cluster__,
                    reservations: reservations__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.Inventory", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Leases {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.active != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.Leases", len)?;
        if self.active != 0 {
            struct_ser.serialize_field("active", &self.active)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Leases {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "active",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Active,
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
                            "active" => Ok(GeneratedField::Active),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Leases;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.Leases")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Leases, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut active__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Leases {
                    active: active__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.Leases", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ManifestStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.deployments != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.ManifestStatus", len)?;
        if self.deployments != 0 {
            struct_ser.serialize_field("deployments", &self.deployments)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ManifestStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "deployments",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Deployments,
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
                            "deployments" => Ok(GeneratedField::Deployments),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ManifestStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.ManifestStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ManifestStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut deployments__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Deployments => {
                            if deployments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deployments"));
                            }
                            deployments__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ManifestStatus {
                    deployments: deployments__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.ManifestStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Reservations {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.pending.is_some() {
            len += 1;
        }
        if self.active.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.Reservations", len)?;
        if let Some(v) = self.pending.as_ref() {
            struct_ser.serialize_field("pending", v)?;
        }
        if let Some(v) = self.active.as_ref() {
            struct_ser.serialize_field("active", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Reservations {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "pending",
            "active",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Pending,
            Active,
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
                            "pending" => Ok(GeneratedField::Pending),
                            "active" => Ok(GeneratedField::Active),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Reservations;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.Reservations")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Reservations, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut pending__ = None;
                let mut active__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Pending => {
                            if pending__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pending"));
                            }
                            pending__ = map_.next_value()?;
                        }
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Reservations {
                    pending: pending__,
                    active: active__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.Reservations", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ReservationsMetric {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.count != 0 {
            len += 1;
        }
        if self.resources.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.ReservationsMetric", len)?;
        if self.count != 0 {
            struct_ser.serialize_field("count", &self.count)?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ReservationsMetric {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "count",
            "resources",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Count,
            Resources,
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
                            "count" => Ok(GeneratedField::Count),
                            "resources" => Ok(GeneratedField::Resources),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ReservationsMetric;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.ReservationsMetric")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ReservationsMetric, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut count__ = None;
                let mut resources__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Count => {
                            if count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("count"));
                            }
                            count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ReservationsMetric {
                    count: count__.unwrap_or_default(),
                    resources: resources__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.ReservationsMetric", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ResourcesMetric {
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
        if !self.storage.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.ResourcesMetric", len)?;
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
        if !self.storage.is_empty() {
            struct_ser.serialize_field("storage", &self.storage)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ResourcesMetric {
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
            "storage",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Cpu,
            Memory,
            Gpu,
            EphemeralStorage,
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
                            "cpu" => Ok(GeneratedField::Cpu),
                            "memory" => Ok(GeneratedField::Memory),
                            "gpu" => Ok(GeneratedField::Gpu),
                            "ephemeralStorage" | "ephemeral_storage" => Ok(GeneratedField::EphemeralStorage),
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
            type Value = ResourcesMetric;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.ResourcesMetric")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ResourcesMetric, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cpu__ = None;
                let mut memory__ = None;
                let mut gpu__ = None;
                let mut ephemeral_storage__ = None;
                let mut storage__ = None;
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
                        GeneratedField::Storage => {
                            if storage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storage"));
                            }
                            storage__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(ResourcesMetric {
                    cpu: cpu__,
                    memory: memory__,
                    gpu: gpu__,
                    ephemeral_storage: ephemeral_storage__,
                    storage: storage__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.ResourcesMetric", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Status {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.errors.is_empty() {
            len += 1;
        }
        if self.cluster.is_some() {
            len += 1;
        }
        if self.bid_engine.is_some() {
            len += 1;
        }
        if self.manifest.is_some() {
            len += 1;
        }
        if !self.public_hostnames.is_empty() {
            len += 1;
        }
        if self.timestamp.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.v1.Status", len)?;
        if !self.errors.is_empty() {
            struct_ser.serialize_field("errors", &self.errors)?;
        }
        if let Some(v) = self.cluster.as_ref() {
            struct_ser.serialize_field("cluster", v)?;
        }
        if let Some(v) = self.bid_engine.as_ref() {
            struct_ser.serialize_field("bidEngine", v)?;
        }
        if let Some(v) = self.manifest.as_ref() {
            struct_ser.serialize_field("manifest", v)?;
        }
        if !self.public_hostnames.is_empty() {
            struct_ser.serialize_field("publicHostnames", &self.public_hostnames)?;
        }
        if let Some(v) = self.timestamp.as_ref() {
            struct_ser.serialize_field("timestamp", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Status {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "errors",
            "cluster",
            "bid_engine",
            "bidEngine",
            "manifest",
            "public_hostnames",
            "publicHostnames",
            "timestamp",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Errors,
            Cluster,
            BidEngine,
            Manifest,
            PublicHostnames,
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
                            "errors" => Ok(GeneratedField::Errors),
                            "cluster" => Ok(GeneratedField::Cluster),
                            "bidEngine" | "bid_engine" => Ok(GeneratedField::BidEngine),
                            "manifest" => Ok(GeneratedField::Manifest),
                            "publicHostnames" | "public_hostnames" => Ok(GeneratedField::PublicHostnames),
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
            type Value = Status;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.v1.Status")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Status, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut errors__ = None;
                let mut cluster__ = None;
                let mut bid_engine__ = None;
                let mut manifest__ = None;
                let mut public_hostnames__ = None;
                let mut timestamp__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Errors => {
                            if errors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("errors"));
                            }
                            errors__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Cluster => {
                            if cluster__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cluster"));
                            }
                            cluster__ = map_.next_value()?;
                        }
                        GeneratedField::BidEngine => {
                            if bid_engine__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bidEngine"));
                            }
                            bid_engine__ = map_.next_value()?;
                        }
                        GeneratedField::Manifest => {
                            if manifest__.is_some() {
                                return Err(serde::de::Error::duplicate_field("manifest"));
                            }
                            manifest__ = map_.next_value()?;
                        }
                        GeneratedField::PublicHostnames => {
                            if public_hostnames__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicHostnames"));
                            }
                            public_hostnames__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Status {
                    errors: errors__.unwrap_or_default(),
                    cluster: cluster__,
                    bid_engine: bid_engine__,
                    manifest: manifest__,
                    public_hostnames: public_hostnames__.unwrap_or_default(),
                    timestamp: timestamp__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.v1.Status", FIELDS, GeneratedVisitor)
    }
}

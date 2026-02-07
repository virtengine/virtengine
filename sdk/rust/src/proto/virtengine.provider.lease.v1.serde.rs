// @generated
impl serde::Serialize for ForwarderPortStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.host.is_empty() {
            len += 1;
        }
        if self.port != 0 {
            len += 1;
        }
        if self.external_port != 0 {
            len += 1;
        }
        if !self.proto.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ForwarderPortStatus", len)?;
        if !self.host.is_empty() {
            struct_ser.serialize_field("host", &self.host)?;
        }
        if self.port != 0 {
            struct_ser.serialize_field("port", &self.port)?;
        }
        if self.external_port != 0 {
            struct_ser.serialize_field("externalPort", &self.external_port)?;
        }
        if !self.proto.is_empty() {
            struct_ser.serialize_field("proto", &self.proto)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ForwarderPortStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "host",
            "port",
            "external_port",
            "externalPort",
            "proto",
            "name",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Host,
            Port,
            ExternalPort,
            Proto,
            Name,
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
                            "host" => Ok(GeneratedField::Host),
                            "port" => Ok(GeneratedField::Port),
                            "externalPort" | "external_port" => Ok(GeneratedField::ExternalPort),
                            "proto" => Ok(GeneratedField::Proto),
                            "name" => Ok(GeneratedField::Name),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ForwarderPortStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ForwarderPortStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ForwarderPortStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut host__ = None;
                let mut port__ = None;
                let mut external_port__ = None;
                let mut proto__ = None;
                let mut name__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Host => {
                            if host__.is_some() {
                                return Err(serde::de::Error::duplicate_field("host"));
                            }
                            host__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Port => {
                            if port__.is_some() {
                                return Err(serde::de::Error::duplicate_field("port"));
                            }
                            port__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExternalPort => {
                            if external_port__.is_some() {
                                return Err(serde::de::Error::duplicate_field("externalPort"));
                            }
                            external_port__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Proto => {
                            if proto__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proto"));
                            }
                            proto__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ForwarderPortStatus {
                    host: host__.unwrap_or_default(),
                    port: port__.unwrap_or_default(),
                    external_port: external_port__.unwrap_or_default(),
                    proto: proto__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ForwarderPortStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LeaseIpStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.port != 0 {
            len += 1;
        }
        if self.external_port != 0 {
            len += 1;
        }
        if !self.protocol.is_empty() {
            len += 1;
        }
        if !self.ip.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.LeaseIPStatus", len)?;
        if self.port != 0 {
            struct_ser.serialize_field("port", &self.port)?;
        }
        if self.external_port != 0 {
            struct_ser.serialize_field("externalPort", &self.external_port)?;
        }
        if !self.protocol.is_empty() {
            struct_ser.serialize_field("protocol", &self.protocol)?;
        }
        if !self.ip.is_empty() {
            struct_ser.serialize_field("ip", &self.ip)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LeaseIpStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "port",
            "external_port",
            "externalPort",
            "protocol",
            "ip",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Port,
            ExternalPort,
            Protocol,
            Ip,
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
                            "port" => Ok(GeneratedField::Port),
                            "externalPort" | "external_port" => Ok(GeneratedField::ExternalPort),
                            "protocol" => Ok(GeneratedField::Protocol),
                            "ip" => Ok(GeneratedField::Ip),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LeaseIpStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.LeaseIPStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LeaseIpStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut port__ = None;
                let mut external_port__ = None;
                let mut protocol__ = None;
                let mut ip__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Port => {
                            if port__.is_some() {
                                return Err(serde::de::Error::duplicate_field("port"));
                            }
                            port__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExternalPort => {
                            if external_port__.is_some() {
                                return Err(serde::de::Error::duplicate_field("externalPort"));
                            }
                            external_port__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Protocol => {
                            if protocol__.is_some() {
                                return Err(serde::de::Error::duplicate_field("protocol"));
                            }
                            protocol__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Ip => {
                            if ip__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ip"));
                            }
                            ip__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(LeaseIpStatus {
                    port: port__.unwrap_or_default(),
                    external_port: external_port__.unwrap_or_default(),
                    protocol: protocol__.unwrap_or_default(),
                    ip: ip__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.LeaseIPStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LeaseServiceStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.available != 0 {
            len += 1;
        }
        if self.total != 0 {
            len += 1;
        }
        if !self.uris.is_empty() {
            len += 1;
        }
        if self.observed_generation != 0 {
            len += 1;
        }
        if self.replicas != 0 {
            len += 1;
        }
        if self.updated_replicas != 0 {
            len += 1;
        }
        if self.ready_replicas != 0 {
            len += 1;
        }
        if self.available_replicas != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.LeaseServiceStatus", len)?;
        if self.available != 0 {
            struct_ser.serialize_field("available", &self.available)?;
        }
        if self.total != 0 {
            struct_ser.serialize_field("total", &self.total)?;
        }
        if !self.uris.is_empty() {
            struct_ser.serialize_field("uris", &self.uris)?;
        }
        if self.observed_generation != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("observedGeneration", ToString::to_string(&self.observed_generation).as_str())?;
        }
        if self.replicas != 0 {
            struct_ser.serialize_field("replicas", &self.replicas)?;
        }
        if self.updated_replicas != 0 {
            struct_ser.serialize_field("updatedReplicas", &self.updated_replicas)?;
        }
        if self.ready_replicas != 0 {
            struct_ser.serialize_field("readyReplicas", &self.ready_replicas)?;
        }
        if self.available_replicas != 0 {
            struct_ser.serialize_field("availableReplicas", &self.available_replicas)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LeaseServiceStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "available",
            "total",
            "uris",
            "observed_generation",
            "observedGeneration",
            "replicas",
            "updated_replicas",
            "updatedReplicas",
            "ready_replicas",
            "readyReplicas",
            "available_replicas",
            "availableReplicas",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Available,
            Total,
            Uris,
            ObservedGeneration,
            Replicas,
            UpdatedReplicas,
            ReadyReplicas,
            AvailableReplicas,
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
                            "available" => Ok(GeneratedField::Available),
                            "total" => Ok(GeneratedField::Total),
                            "uris" => Ok(GeneratedField::Uris),
                            "observedGeneration" | "observed_generation" => Ok(GeneratedField::ObservedGeneration),
                            "replicas" => Ok(GeneratedField::Replicas),
                            "updatedReplicas" | "updated_replicas" => Ok(GeneratedField::UpdatedReplicas),
                            "readyReplicas" | "ready_replicas" => Ok(GeneratedField::ReadyReplicas),
                            "availableReplicas" | "available_replicas" => Ok(GeneratedField::AvailableReplicas),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LeaseServiceStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.LeaseServiceStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LeaseServiceStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut available__ = None;
                let mut total__ = None;
                let mut uris__ = None;
                let mut observed_generation__ = None;
                let mut replicas__ = None;
                let mut updated_replicas__ = None;
                let mut ready_replicas__ = None;
                let mut available_replicas__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Available => {
                            if available__.is_some() {
                                return Err(serde::de::Error::duplicate_field("available"));
                            }
                            available__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Total => {
                            if total__.is_some() {
                                return Err(serde::de::Error::duplicate_field("total"));
                            }
                            total__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Uris => {
                            if uris__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uris"));
                            }
                            uris__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ObservedGeneration => {
                            if observed_generation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("observedGeneration"));
                            }
                            observed_generation__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Replicas => {
                            if replicas__.is_some() {
                                return Err(serde::de::Error::duplicate_field("replicas"));
                            }
                            replicas__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpdatedReplicas => {
                            if updated_replicas__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedReplicas"));
                            }
                            updated_replicas__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReadyReplicas => {
                            if ready_replicas__.is_some() {
                                return Err(serde::de::Error::duplicate_field("readyReplicas"));
                            }
                            ready_replicas__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AvailableReplicas => {
                            if available_replicas__.is_some() {
                                return Err(serde::de::Error::duplicate_field("availableReplicas"));
                            }
                            available_replicas__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(LeaseServiceStatus {
                    available: available__.unwrap_or_default(),
                    total: total__.unwrap_or_default(),
                    uris: uris__.unwrap_or_default(),
                    observed_generation: observed_generation__.unwrap_or_default(),
                    replicas: replicas__.unwrap_or_default(),
                    updated_replicas: updated_replicas__.unwrap_or_default(),
                    ready_replicas: ready_replicas__.unwrap_or_default(),
                    available_replicas: available_replicas__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.LeaseServiceStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SendManifestRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.lease_id.is_some() {
            len += 1;
        }
        if !self.manifest.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.SendManifestRequest", len)?;
        if let Some(v) = self.lease_id.as_ref() {
            struct_ser.serialize_field("leaseId", v)?;
        }
        if !self.manifest.is_empty() {
            struct_ser.serialize_field("manifest", &self.manifest)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SendManifestRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "lease_id",
            "leaseId",
            "manifest",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            LeaseId,
            Manifest,
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
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            "manifest" => Ok(GeneratedField::Manifest),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SendManifestRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.SendManifestRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SendManifestRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut lease_id__ = None;
                let mut manifest__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = map_.next_value()?;
                        }
                        GeneratedField::Manifest => {
                            if manifest__.is_some() {
                                return Err(serde::de::Error::duplicate_field("manifest"));
                            }
                            manifest__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(SendManifestRequest {
                    lease_id: lease_id__,
                    manifest: manifest__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.SendManifestRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SendManifestResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.SendManifestResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SendManifestResponse {
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
            type Value = SendManifestResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.SendManifestResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SendManifestResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(SendManifestResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.SendManifestResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceLogs {
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
        if !self.logs.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ServiceLogs", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.logs.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("logs", pbjson::private::base64::encode(&self.logs).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceLogs {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "logs",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Logs,
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
                            "logs" => Ok(GeneratedField::Logs),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceLogs;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ServiceLogs")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceLogs, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut logs__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Logs => {
                            if logs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("logs"));
                            }
                            logs__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ServiceLogs {
                    name: name__.unwrap_or_default(),
                    logs: logs__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ServiceLogs", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceLogsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.lease_id.is_some() {
            len += 1;
        }
        if !self.services.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ServiceLogsRequest", len)?;
        if let Some(v) = self.lease_id.as_ref() {
            struct_ser.serialize_field("leaseId", v)?;
        }
        if !self.services.is_empty() {
            struct_ser.serialize_field("services", &self.services)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceLogsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "lease_id",
            "leaseId",
            "services",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            LeaseId,
            Services,
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
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            "services" => Ok(GeneratedField::Services),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceLogsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ServiceLogsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceLogsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut lease_id__ = None;
                let mut services__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = map_.next_value()?;
                        }
                        GeneratedField::Services => {
                            if services__.is_some() {
                                return Err(serde::de::Error::duplicate_field("services"));
                            }
                            services__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ServiceLogsRequest {
                    lease_id: lease_id__,
                    services: services__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ServiceLogsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceLogsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.services.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ServiceLogsResponse", len)?;
        if !self.services.is_empty() {
            struct_ser.serialize_field("services", &self.services)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceLogsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "services",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Services,
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
                            "services" => Ok(GeneratedField::Services),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceLogsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ServiceLogsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceLogsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut services__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Services => {
                            if services__.is_some() {
                                return Err(serde::de::Error::duplicate_field("services"));
                            }
                            services__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ServiceLogsResponse {
                    services: services__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ServiceLogsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceStatus {
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
        if self.status.is_some() {
            len += 1;
        }
        if !self.ports.is_empty() {
            len += 1;
        }
        if !self.ips.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ServiceStatus", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if let Some(v) = self.status.as_ref() {
            struct_ser.serialize_field("status", v)?;
        }
        if !self.ports.is_empty() {
            struct_ser.serialize_field("ports", &self.ports)?;
        }
        if !self.ips.is_empty() {
            struct_ser.serialize_field("ips", &self.ips)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "status",
            "ports",
            "ips",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Status,
            Ports,
            Ips,
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
                            "status" => Ok(GeneratedField::Status),
                            "ports" => Ok(GeneratedField::Ports),
                            "ips" => Ok(GeneratedField::Ips),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ServiceStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut status__ = None;
                let mut ports__ = None;
                let mut ips__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = map_.next_value()?;
                        }
                        GeneratedField::Ports => {
                            if ports__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ports"));
                            }
                            ports__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Ips => {
                            if ips__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ips"));
                            }
                            ips__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ServiceStatus {
                    name: name__.unwrap_or_default(),
                    status: status__,
                    ports: ports__.unwrap_or_default(),
                    ips: ips__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ServiceStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceStatusRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.lease_id.is_some() {
            len += 1;
        }
        if !self.services.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ServiceStatusRequest", len)?;
        if let Some(v) = self.lease_id.as_ref() {
            struct_ser.serialize_field("leaseId", v)?;
        }
        if !self.services.is_empty() {
            struct_ser.serialize_field("services", &self.services)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceStatusRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "lease_id",
            "leaseId",
            "services",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            LeaseId,
            Services,
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
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            "services" => Ok(GeneratedField::Services),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceStatusRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ServiceStatusRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceStatusRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut lease_id__ = None;
                let mut services__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = map_.next_value()?;
                        }
                        GeneratedField::Services => {
                            if services__.is_some() {
                                return Err(serde::de::Error::duplicate_field("services"));
                            }
                            services__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ServiceStatusRequest {
                    lease_id: lease_id__,
                    services: services__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ServiceStatusRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceStatusResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.services.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ServiceStatusResponse", len)?;
        if !self.services.is_empty() {
            struct_ser.serialize_field("services", &self.services)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceStatusResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "services",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Services,
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
                            "services" => Ok(GeneratedField::Services),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ServiceStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut services__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Services => {
                            if services__.is_some() {
                                return Err(serde::de::Error::duplicate_field("services"));
                            }
                            services__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ServiceStatusResponse {
                    services: services__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ServiceStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ShellRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.lease_id.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.provider.lease.v1.ShellRequest", len)?;
        if let Some(v) = self.lease_id.as_ref() {
            struct_ser.serialize_field("leaseId", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ShellRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "lease_id",
            "leaseId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            LeaseId,
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
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ShellRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.provider.lease.v1.ShellRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ShellRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut lease_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ShellRequest {
                    lease_id: lease_id__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.provider.lease.v1.ShellRequest", FIELDS, GeneratedVisitor)
    }
}

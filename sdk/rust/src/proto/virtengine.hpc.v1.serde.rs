// @generated
impl serde::Serialize for ClusterCandidate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.region.is_empty() {
            len += 1;
        }
        if self.avg_latency_ms != 0 {
            len += 1;
        }
        if self.available_nodes != 0 {
            len += 1;
        }
        if !self.latency_score.is_empty() {
            len += 1;
        }
        if !self.capacity_score.is_empty() {
            len += 1;
        }
        if !self.combined_score.is_empty() {
            len += 1;
        }
        if self.eligible {
            len += 1;
        }
        if !self.ineligibility_reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.ClusterCandidate", len)?;
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.region.is_empty() {
            struct_ser.serialize_field("region", &self.region)?;
        }
        if self.avg_latency_ms != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("avgLatencyMs", ToString::to_string(&self.avg_latency_ms).as_str())?;
        }
        if self.available_nodes != 0 {
            struct_ser.serialize_field("availableNodes", &self.available_nodes)?;
        }
        if !self.latency_score.is_empty() {
            struct_ser.serialize_field("latencyScore", &self.latency_score)?;
        }
        if !self.capacity_score.is_empty() {
            struct_ser.serialize_field("capacityScore", &self.capacity_score)?;
        }
        if !self.combined_score.is_empty() {
            struct_ser.serialize_field("combinedScore", &self.combined_score)?;
        }
        if self.eligible {
            struct_ser.serialize_field("eligible", &self.eligible)?;
        }
        if !self.ineligibility_reason.is_empty() {
            struct_ser.serialize_field("ineligibilityReason", &self.ineligibility_reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ClusterCandidate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster_id",
            "clusterId",
            "region",
            "avg_latency_ms",
            "avgLatencyMs",
            "available_nodes",
            "availableNodes",
            "latency_score",
            "latencyScore",
            "capacity_score",
            "capacityScore",
            "combined_score",
            "combinedScore",
            "eligible",
            "ineligibility_reason",
            "ineligibilityReason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClusterId,
            Region,
            AvgLatencyMs,
            AvailableNodes,
            LatencyScore,
            CapacityScore,
            CombinedScore,
            Eligible,
            IneligibilityReason,
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
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "region" => Ok(GeneratedField::Region),
                            "avgLatencyMs" | "avg_latency_ms" => Ok(GeneratedField::AvgLatencyMs),
                            "availableNodes" | "available_nodes" => Ok(GeneratedField::AvailableNodes),
                            "latencyScore" | "latency_score" => Ok(GeneratedField::LatencyScore),
                            "capacityScore" | "capacity_score" => Ok(GeneratedField::CapacityScore),
                            "combinedScore" | "combined_score" => Ok(GeneratedField::CombinedScore),
                            "eligible" => Ok(GeneratedField::Eligible),
                            "ineligibilityReason" | "ineligibility_reason" => Ok(GeneratedField::IneligibilityReason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ClusterCandidate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.ClusterCandidate")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ClusterCandidate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster_id__ = None;
                let mut region__ = None;
                let mut avg_latency_ms__ = None;
                let mut available_nodes__ = None;
                let mut latency_score__ = None;
                let mut capacity_score__ = None;
                let mut combined_score__ = None;
                let mut eligible__ = None;
                let mut ineligibility_reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Region => {
                            if region__.is_some() {
                                return Err(serde::de::Error::duplicate_field("region"));
                            }
                            region__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AvgLatencyMs => {
                            if avg_latency_ms__.is_some() {
                                return Err(serde::de::Error::duplicate_field("avgLatencyMs"));
                            }
                            avg_latency_ms__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AvailableNodes => {
                            if available_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("availableNodes"));
                            }
                            available_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LatencyScore => {
                            if latency_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latencyScore"));
                            }
                            latency_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CapacityScore => {
                            if capacity_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("capacityScore"));
                            }
                            capacity_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CombinedScore => {
                            if combined_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("combinedScore"));
                            }
                            combined_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Eligible => {
                            if eligible__.is_some() {
                                return Err(serde::de::Error::duplicate_field("eligible"));
                            }
                            eligible__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IneligibilityReason => {
                            if ineligibility_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ineligibilityReason"));
                            }
                            ineligibility_reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ClusterCandidate {
                    cluster_id: cluster_id__.unwrap_or_default(),
                    region: region__.unwrap_or_default(),
                    avg_latency_ms: avg_latency_ms__.unwrap_or_default(),
                    available_nodes: available_nodes__.unwrap_or_default(),
                    latency_score: latency_score__.unwrap_or_default(),
                    capacity_score: capacity_score__.unwrap_or_default(),
                    combined_score: combined_score__.unwrap_or_default(),
                    eligible: eligible__.unwrap_or_default(),
                    ineligibility_reason: ineligibility_reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.ClusterCandidate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ClusterMetadata {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.total_cpu_cores != 0 {
            len += 1;
        }
        if self.total_memory_gb != 0 {
            len += 1;
        }
        if self.total_gpus != 0 {
            len += 1;
        }
        if !self.gpu_types.is_empty() {
            len += 1;
        }
        if !self.interconnect_type.is_empty() {
            len += 1;
        }
        if !self.storage_type.is_empty() {
            len += 1;
        }
        if self.total_storage_gb != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.ClusterMetadata", len)?;
        if self.total_cpu_cores != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("totalCpuCores", ToString::to_string(&self.total_cpu_cores).as_str())?;
        }
        if self.total_memory_gb != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("totalMemoryGb", ToString::to_string(&self.total_memory_gb).as_str())?;
        }
        if self.total_gpus != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("totalGpus", ToString::to_string(&self.total_gpus).as_str())?;
        }
        if !self.gpu_types.is_empty() {
            struct_ser.serialize_field("gpuTypes", &self.gpu_types)?;
        }
        if !self.interconnect_type.is_empty() {
            struct_ser.serialize_field("interconnectType", &self.interconnect_type)?;
        }
        if !self.storage_type.is_empty() {
            struct_ser.serialize_field("storageType", &self.storage_type)?;
        }
        if self.total_storage_gb != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("totalStorageGb", ToString::to_string(&self.total_storage_gb).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ClusterMetadata {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "total_cpu_cores",
            "totalCpuCores",
            "total_memory_gb",
            "totalMemoryGb",
            "total_gpus",
            "totalGpus",
            "gpu_types",
            "gpuTypes",
            "interconnect_type",
            "interconnectType",
            "storage_type",
            "storageType",
            "total_storage_gb",
            "totalStorageGb",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TotalCpuCores,
            TotalMemoryGb,
            TotalGpus,
            GpuTypes,
            InterconnectType,
            StorageType,
            TotalStorageGb,
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
                            "totalCpuCores" | "total_cpu_cores" => Ok(GeneratedField::TotalCpuCores),
                            "totalMemoryGb" | "total_memory_gb" => Ok(GeneratedField::TotalMemoryGb),
                            "totalGpus" | "total_gpus" => Ok(GeneratedField::TotalGpus),
                            "gpuTypes" | "gpu_types" => Ok(GeneratedField::GpuTypes),
                            "interconnectType" | "interconnect_type" => Ok(GeneratedField::InterconnectType),
                            "storageType" | "storage_type" => Ok(GeneratedField::StorageType),
                            "totalStorageGb" | "total_storage_gb" => Ok(GeneratedField::TotalStorageGb),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ClusterMetadata;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.ClusterMetadata")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ClusterMetadata, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut total_cpu_cores__ = None;
                let mut total_memory_gb__ = None;
                let mut total_gpus__ = None;
                let mut gpu_types__ = None;
                let mut interconnect_type__ = None;
                let mut storage_type__ = None;
                let mut total_storage_gb__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TotalCpuCores => {
                            if total_cpu_cores__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalCpuCores"));
                            }
                            total_cpu_cores__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TotalMemoryGb => {
                            if total_memory_gb__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalMemoryGb"));
                            }
                            total_memory_gb__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TotalGpus => {
                            if total_gpus__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalGpus"));
                            }
                            total_gpus__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GpuTypes => {
                            if gpu_types__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpuTypes"));
                            }
                            gpu_types__ = Some(map_.next_value()?);
                        }
                        GeneratedField::InterconnectType => {
                            if interconnect_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("interconnectType"));
                            }
                            interconnect_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StorageType => {
                            if storage_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storageType"));
                            }
                            storage_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalStorageGb => {
                            if total_storage_gb__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalStorageGb"));
                            }
                            total_storage_gb__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ClusterMetadata {
                    total_cpu_cores: total_cpu_cores__.unwrap_or_default(),
                    total_memory_gb: total_memory_gb__.unwrap_or_default(),
                    total_gpus: total_gpus__.unwrap_or_default(),
                    gpu_types: gpu_types__.unwrap_or_default(),
                    interconnect_type: interconnect_type__.unwrap_or_default(),
                    storage_type: storage_type__.unwrap_or_default(),
                    total_storage_gb: total_storage_gb__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.ClusterMetadata", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ClusterState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "CLUSTER_STATE_UNSPECIFIED",
            Self::Pending => "CLUSTER_STATE_PENDING",
            Self::Active => "CLUSTER_STATE_ACTIVE",
            Self::Draining => "CLUSTER_STATE_DRAINING",
            Self::Offline => "CLUSTER_STATE_OFFLINE",
            Self::Deregistered => "CLUSTER_STATE_DEREGISTERED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ClusterState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "CLUSTER_STATE_UNSPECIFIED",
            "CLUSTER_STATE_PENDING",
            "CLUSTER_STATE_ACTIVE",
            "CLUSTER_STATE_DRAINING",
            "CLUSTER_STATE_OFFLINE",
            "CLUSTER_STATE_DEREGISTERED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ClusterState;

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
                    "CLUSTER_STATE_UNSPECIFIED" => Ok(ClusterState::Unspecified),
                    "CLUSTER_STATE_PENDING" => Ok(ClusterState::Pending),
                    "CLUSTER_STATE_ACTIVE" => Ok(ClusterState::Active),
                    "CLUSTER_STATE_DRAINING" => Ok(ClusterState::Draining),
                    "CLUSTER_STATE_OFFLINE" => Ok(ClusterState::Offline),
                    "CLUSTER_STATE_DEREGISTERED" => Ok(ClusterState::Deregistered),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for DataReference {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reference_id.is_empty() {
            len += 1;
        }
        if !self.r#type.is_empty() {
            len += 1;
        }
        if !self.uri.is_empty() {
            len += 1;
        }
        if self.encrypted {
            len += 1;
        }
        if !self.checksum.is_empty() {
            len += 1;
        }
        if self.size_bytes != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.DataReference", len)?;
        if !self.reference_id.is_empty() {
            struct_ser.serialize_field("referenceId", &self.reference_id)?;
        }
        if !self.r#type.is_empty() {
            struct_ser.serialize_field("type", &self.r#type)?;
        }
        if !self.uri.is_empty() {
            struct_ser.serialize_field("uri", &self.uri)?;
        }
        if self.encrypted {
            struct_ser.serialize_field("encrypted", &self.encrypted)?;
        }
        if !self.checksum.is_empty() {
            struct_ser.serialize_field("checksum", &self.checksum)?;
        }
        if self.size_bytes != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sizeBytes", ToString::to_string(&self.size_bytes).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DataReference {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reference_id",
            "referenceId",
            "type",
            "uri",
            "encrypted",
            "checksum",
            "size_bytes",
            "sizeBytes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReferenceId,
            Type,
            Uri,
            Encrypted,
            Checksum,
            SizeBytes,
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
                            "referenceId" | "reference_id" => Ok(GeneratedField::ReferenceId),
                            "type" => Ok(GeneratedField::Type),
                            "uri" => Ok(GeneratedField::Uri),
                            "encrypted" => Ok(GeneratedField::Encrypted),
                            "checksum" => Ok(GeneratedField::Checksum),
                            "sizeBytes" | "size_bytes" => Ok(GeneratedField::SizeBytes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DataReference;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.DataReference")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DataReference, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reference_id__ = None;
                let mut r#type__ = None;
                let mut uri__ = None;
                let mut encrypted__ = None;
                let mut checksum__ = None;
                let mut size_bytes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReferenceId => {
                            if reference_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("referenceId"));
                            }
                            reference_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Type => {
                            if r#type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("type"));
                            }
                            r#type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Uri => {
                            if uri__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uri"));
                            }
                            uri__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Encrypted => {
                            if encrypted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encrypted"));
                            }
                            encrypted__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Checksum => {
                            if checksum__.is_some() {
                                return Err(serde::de::Error::duplicate_field("checksum"));
                            }
                            checksum__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SizeBytes => {
                            if size_bytes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sizeBytes"));
                            }
                            size_bytes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(DataReference {
                    reference_id: reference_id__.unwrap_or_default(),
                    r#type: r#type__.unwrap_or_default(),
                    uri: uri__.unwrap_or_default(),
                    encrypted: encrypted__.unwrap_or_default(),
                    checksum: checksum__.unwrap_or_default(),
                    size_bytes: size_bytes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.DataReference", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for DisputeStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "DISPUTE_STATUS_UNSPECIFIED",
            Self::Pending => "DISPUTE_STATUS_PENDING",
            Self::UnderReview => "DISPUTE_STATUS_UNDER_REVIEW",
            Self::Resolved => "DISPUTE_STATUS_RESOLVED",
            Self::Rejected => "DISPUTE_STATUS_REJECTED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for DisputeStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "DISPUTE_STATUS_UNSPECIFIED",
            "DISPUTE_STATUS_PENDING",
            "DISPUTE_STATUS_UNDER_REVIEW",
            "DISPUTE_STATUS_RESOLVED",
            "DISPUTE_STATUS_REJECTED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DisputeStatus;

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
                    "DISPUTE_STATUS_UNSPECIFIED" => Ok(DisputeStatus::Unspecified),
                    "DISPUTE_STATUS_PENDING" => Ok(DisputeStatus::Pending),
                    "DISPUTE_STATUS_UNDER_REVIEW" => Ok(DisputeStatus::UnderReview),
                    "DISPUTE_STATUS_RESOLVED" => Ok(DisputeStatus::Resolved),
                    "DISPUTE_STATUS_REJECTED" => Ok(DisputeStatus::Rejected),
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
        if !self.clusters.is_empty() {
            len += 1;
        }
        if !self.offerings.is_empty() {
            len += 1;
        }
        if !self.jobs.is_empty() {
            len += 1;
        }
        if !self.job_accountings.is_empty() {
            len += 1;
        }
        if !self.node_metadatas.is_empty() {
            len += 1;
        }
        if !self.scheduling_decisions.is_empty() {
            len += 1;
        }
        if !self.hpc_rewards.is_empty() {
            len += 1;
        }
        if !self.disputes.is_empty() {
            len += 1;
        }
        if self.cluster_sequence != 0 {
            len += 1;
        }
        if self.offering_sequence != 0 {
            len += 1;
        }
        if self.job_sequence != 0 {
            len += 1;
        }
        if self.decision_sequence != 0 {
            len += 1;
        }
        if self.dispute_sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.clusters.is_empty() {
            struct_ser.serialize_field("clusters", &self.clusters)?;
        }
        if !self.offerings.is_empty() {
            struct_ser.serialize_field("offerings", &self.offerings)?;
        }
        if !self.jobs.is_empty() {
            struct_ser.serialize_field("jobs", &self.jobs)?;
        }
        if !self.job_accountings.is_empty() {
            struct_ser.serialize_field("jobAccountings", &self.job_accountings)?;
        }
        if !self.node_metadatas.is_empty() {
            struct_ser.serialize_field("nodeMetadatas", &self.node_metadatas)?;
        }
        if !self.scheduling_decisions.is_empty() {
            struct_ser.serialize_field("schedulingDecisions", &self.scheduling_decisions)?;
        }
        if !self.hpc_rewards.is_empty() {
            struct_ser.serialize_field("hpcRewards", &self.hpc_rewards)?;
        }
        if !self.disputes.is_empty() {
            struct_ser.serialize_field("disputes", &self.disputes)?;
        }
        if self.cluster_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("clusterSequence", ToString::to_string(&self.cluster_sequence).as_str())?;
        }
        if self.offering_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("offeringSequence", ToString::to_string(&self.offering_sequence).as_str())?;
        }
        if self.job_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("jobSequence", ToString::to_string(&self.job_sequence).as_str())?;
        }
        if self.decision_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("decisionSequence", ToString::to_string(&self.decision_sequence).as_str())?;
        }
        if self.dispute_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("disputeSequence", ToString::to_string(&self.dispute_sequence).as_str())?;
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
            "clusters",
            "offerings",
            "jobs",
            "job_accountings",
            "jobAccountings",
            "node_metadatas",
            "nodeMetadatas",
            "scheduling_decisions",
            "schedulingDecisions",
            "hpc_rewards",
            "hpcRewards",
            "disputes",
            "cluster_sequence",
            "clusterSequence",
            "offering_sequence",
            "offeringSequence",
            "job_sequence",
            "jobSequence",
            "decision_sequence",
            "decisionSequence",
            "dispute_sequence",
            "disputeSequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            Clusters,
            Offerings,
            Jobs,
            JobAccountings,
            NodeMetadatas,
            SchedulingDecisions,
            HpcRewards,
            Disputes,
            ClusterSequence,
            OfferingSequence,
            JobSequence,
            DecisionSequence,
            DisputeSequence,
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
                            "clusters" => Ok(GeneratedField::Clusters),
                            "offerings" => Ok(GeneratedField::Offerings),
                            "jobs" => Ok(GeneratedField::Jobs),
                            "jobAccountings" | "job_accountings" => Ok(GeneratedField::JobAccountings),
                            "nodeMetadatas" | "node_metadatas" => Ok(GeneratedField::NodeMetadatas),
                            "schedulingDecisions" | "scheduling_decisions" => Ok(GeneratedField::SchedulingDecisions),
                            "hpcRewards" | "hpc_rewards" => Ok(GeneratedField::HpcRewards),
                            "disputes" => Ok(GeneratedField::Disputes),
                            "clusterSequence" | "cluster_sequence" => Ok(GeneratedField::ClusterSequence),
                            "offeringSequence" | "offering_sequence" => Ok(GeneratedField::OfferingSequence),
                            "jobSequence" | "job_sequence" => Ok(GeneratedField::JobSequence),
                            "decisionSequence" | "decision_sequence" => Ok(GeneratedField::DecisionSequence),
                            "disputeSequence" | "dispute_sequence" => Ok(GeneratedField::DisputeSequence),
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
                formatter.write_str("struct virtengine.hpc.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut clusters__ = None;
                let mut offerings__ = None;
                let mut jobs__ = None;
                let mut job_accountings__ = None;
                let mut node_metadatas__ = None;
                let mut scheduling_decisions__ = None;
                let mut hpc_rewards__ = None;
                let mut disputes__ = None;
                let mut cluster_sequence__ = None;
                let mut offering_sequence__ = None;
                let mut job_sequence__ = None;
                let mut decision_sequence__ = None;
                let mut dispute_sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Clusters => {
                            if clusters__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusters"));
                            }
                            clusters__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Offerings => {
                            if offerings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offerings"));
                            }
                            offerings__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Jobs => {
                            if jobs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobs"));
                            }
                            jobs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobAccountings => {
                            if job_accountings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobAccountings"));
                            }
                            job_accountings__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NodeMetadatas => {
                            if node_metadatas__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeMetadatas"));
                            }
                            node_metadatas__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SchedulingDecisions => {
                            if scheduling_decisions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("schedulingDecisions"));
                            }
                            scheduling_decisions__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HpcRewards => {
                            if hpc_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hpcRewards"));
                            }
                            hpc_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Disputes => {
                            if disputes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputes"));
                            }
                            disputes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterSequence => {
                            if cluster_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterSequence"));
                            }
                            cluster_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OfferingSequence => {
                            if offering_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringSequence"));
                            }
                            offering_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JobSequence => {
                            if job_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobSequence"));
                            }
                            job_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DecisionSequence => {
                            if decision_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("decisionSequence"));
                            }
                            decision_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DisputeSequence => {
                            if dispute_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeSequence"));
                            }
                            dispute_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    clusters: clusters__.unwrap_or_default(),
                    offerings: offerings__.unwrap_or_default(),
                    jobs: jobs__.unwrap_or_default(),
                    job_accountings: job_accountings__.unwrap_or_default(),
                    node_metadatas: node_metadatas__.unwrap_or_default(),
                    scheduling_decisions: scheduling_decisions__.unwrap_or_default(),
                    hpc_rewards: hpc_rewards__.unwrap_or_default(),
                    disputes: disputes__.unwrap_or_default(),
                    cluster_sequence: cluster_sequence__.unwrap_or_default(),
                    offering_sequence: offering_sequence__.unwrap_or_default(),
                    job_sequence: job_sequence__.unwrap_or_default(),
                    decision_sequence: decision_sequence__.unwrap_or_default(),
                    dispute_sequence: dispute_sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcCluster {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.state != 0 {
            len += 1;
        }
        if !self.partitions.is_empty() {
            len += 1;
        }
        if self.total_nodes != 0 {
            len += 1;
        }
        if self.available_nodes != 0 {
            len += 1;
        }
        if !self.region.is_empty() {
            len += 1;
        }
        if self.cluster_metadata.is_some() {
            len += 1;
        }
        if !self.slurm_version.is_empty() {
            len += 1;
        }
        if !self.kubernetes_cluster_id.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCCluster", len)?;
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.state != 0 {
            let v = ClusterState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.partitions.is_empty() {
            struct_ser.serialize_field("partitions", &self.partitions)?;
        }
        if self.total_nodes != 0 {
            struct_ser.serialize_field("totalNodes", &self.total_nodes)?;
        }
        if self.available_nodes != 0 {
            struct_ser.serialize_field("availableNodes", &self.available_nodes)?;
        }
        if !self.region.is_empty() {
            struct_ser.serialize_field("region", &self.region)?;
        }
        if let Some(v) = self.cluster_metadata.as_ref() {
            struct_ser.serialize_field("clusterMetadata", v)?;
        }
        if !self.slurm_version.is_empty() {
            struct_ser.serialize_field("slurmVersion", &self.slurm_version)?;
        }
        if !self.kubernetes_cluster_id.is_empty() {
            struct_ser.serialize_field("kubernetesClusterId", &self.kubernetes_cluster_id)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcCluster {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster_id",
            "clusterId",
            "provider_address",
            "providerAddress",
            "name",
            "description",
            "state",
            "partitions",
            "total_nodes",
            "totalNodes",
            "available_nodes",
            "availableNodes",
            "region",
            "cluster_metadata",
            "clusterMetadata",
            "slurm_version",
            "slurmVersion",
            "kubernetes_cluster_id",
            "kubernetesClusterId",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClusterId,
            ProviderAddress,
            Name,
            Description,
            State,
            Partitions,
            TotalNodes,
            AvailableNodes,
            Region,
            ClusterMetadata,
            SlurmVersion,
            KubernetesClusterId,
            CreatedAt,
            UpdatedAt,
            BlockHeight,
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
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "state" => Ok(GeneratedField::State),
                            "partitions" => Ok(GeneratedField::Partitions),
                            "totalNodes" | "total_nodes" => Ok(GeneratedField::TotalNodes),
                            "availableNodes" | "available_nodes" => Ok(GeneratedField::AvailableNodes),
                            "region" => Ok(GeneratedField::Region),
                            "clusterMetadata" | "cluster_metadata" => Ok(GeneratedField::ClusterMetadata),
                            "slurmVersion" | "slurm_version" => Ok(GeneratedField::SlurmVersion),
                            "kubernetesClusterId" | "kubernetes_cluster_id" => Ok(GeneratedField::KubernetesClusterId),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcCluster;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCCluster")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcCluster, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster_id__ = None;
                let mut provider_address__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut state__ = None;
                let mut partitions__ = None;
                let mut total_nodes__ = None;
                let mut available_nodes__ = None;
                let mut region__ = None;
                let mut cluster_metadata__ = None;
                let mut slurm_version__ = None;
                let mut kubernetes_cluster_id__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<ClusterState>()? as i32);
                        }
                        GeneratedField::Partitions => {
                            if partitions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("partitions"));
                            }
                            partitions__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalNodes => {
                            if total_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalNodes"));
                            }
                            total_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AvailableNodes => {
                            if available_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("availableNodes"));
                            }
                            available_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Region => {
                            if region__.is_some() {
                                return Err(serde::de::Error::duplicate_field("region"));
                            }
                            region__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterMetadata => {
                            if cluster_metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterMetadata"));
                            }
                            cluster_metadata__ = map_.next_value()?;
                        }
                        GeneratedField::SlurmVersion => {
                            if slurm_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slurmVersion"));
                            }
                            slurm_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::KubernetesClusterId => {
                            if kubernetes_cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("kubernetesClusterId"));
                            }
                            kubernetes_cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(HpcCluster {
                    cluster_id: cluster_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    partitions: partitions__.unwrap_or_default(),
                    total_nodes: total_nodes__.unwrap_or_default(),
                    available_nodes: available_nodes__.unwrap_or_default(),
                    region: region__.unwrap_or_default(),
                    cluster_metadata: cluster_metadata__,
                    slurm_version: slurm_version__.unwrap_or_default(),
                    kubernetes_cluster_id: kubernetes_cluster_id__.unwrap_or_default(),
                    created_at: created_at__,
                    updated_at: updated_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCCluster", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcDispute {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.dispute_id.is_empty() {
            len += 1;
        }
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.reward_id.is_empty() {
            len += 1;
        }
        if !self.disputer_address.is_empty() {
            len += 1;
        }
        if !self.dispute_type.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.resolution.is_empty() {
            len += 1;
        }
        if !self.resolver_address.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.resolved_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCDispute", len)?;
        if !self.dispute_id.is_empty() {
            struct_ser.serialize_field("disputeId", &self.dispute_id)?;
        }
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.reward_id.is_empty() {
            struct_ser.serialize_field("rewardId", &self.reward_id)?;
        }
        if !self.disputer_address.is_empty() {
            struct_ser.serialize_field("disputerAddress", &self.disputer_address)?;
        }
        if !self.dispute_type.is_empty() {
            struct_ser.serialize_field("disputeType", &self.dispute_type)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        if self.status != 0 {
            let v = DisputeStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.resolution.is_empty() {
            struct_ser.serialize_field("resolution", &self.resolution)?;
        }
        if !self.resolver_address.is_empty() {
            struct_ser.serialize_field("resolverAddress", &self.resolver_address)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.resolved_at.as_ref() {
            struct_ser.serialize_field("resolvedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcDispute {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "dispute_id",
            "disputeId",
            "job_id",
            "jobId",
            "reward_id",
            "rewardId",
            "disputer_address",
            "disputerAddress",
            "dispute_type",
            "disputeType",
            "reason",
            "evidence",
            "status",
            "resolution",
            "resolver_address",
            "resolverAddress",
            "created_at",
            "createdAt",
            "resolved_at",
            "resolvedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DisputeId,
            JobId,
            RewardId,
            DisputerAddress,
            DisputeType,
            Reason,
            Evidence,
            Status,
            Resolution,
            ResolverAddress,
            CreatedAt,
            ResolvedAt,
            BlockHeight,
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
                            "disputeId" | "dispute_id" => Ok(GeneratedField::DisputeId),
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "rewardId" | "reward_id" => Ok(GeneratedField::RewardId),
                            "disputerAddress" | "disputer_address" => Ok(GeneratedField::DisputerAddress),
                            "disputeType" | "dispute_type" => Ok(GeneratedField::DisputeType),
                            "reason" => Ok(GeneratedField::Reason),
                            "evidence" => Ok(GeneratedField::Evidence),
                            "status" => Ok(GeneratedField::Status),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "resolverAddress" | "resolver_address" => Ok(GeneratedField::ResolverAddress),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "resolvedAt" | "resolved_at" => Ok(GeneratedField::ResolvedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcDispute;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCDispute")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcDispute, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut dispute_id__ = None;
                let mut job_id__ = None;
                let mut reward_id__ = None;
                let mut disputer_address__ = None;
                let mut dispute_type__ = None;
                let mut reason__ = None;
                let mut evidence__ = None;
                let mut status__ = None;
                let mut resolution__ = None;
                let mut resolver_address__ = None;
                let mut created_at__ = None;
                let mut resolved_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DisputeId => {
                            if dispute_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeId"));
                            }
                            dispute_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RewardId => {
                            if reward_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardId"));
                            }
                            reward_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DisputerAddress => {
                            if disputer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputerAddress"));
                            }
                            disputer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DisputeType => {
                            if dispute_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeType"));
                            }
                            dispute_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<DisputeStatus>()? as i32);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResolverAddress => {
                            if resolver_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolverAddress"));
                            }
                            resolver_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
                        }
                        GeneratedField::ResolvedAt => {
                            if resolved_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolvedAt"));
                            }
                            resolved_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(HpcDispute {
                    dispute_id: dispute_id__.unwrap_or_default(),
                    job_id: job_id__.unwrap_or_default(),
                    reward_id: reward_id__.unwrap_or_default(),
                    disputer_address: disputer_address__.unwrap_or_default(),
                    dispute_type: dispute_type__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    resolver_address: resolver_address__.unwrap_or_default(),
                    created_at: created_at__,
                    resolved_at: resolved_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCDispute", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcJob {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.offering_id.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.customer_address.is_empty() {
            len += 1;
        }
        if !self.slurm_job_id.is_empty() {
            len += 1;
        }
        if self.state != 0 {
            len += 1;
        }
        if !self.queue_name.is_empty() {
            len += 1;
        }
        if self.workload_spec.is_some() {
            len += 1;
        }
        if self.resources.is_some() {
            len += 1;
        }
        if !self.data_references.is_empty() {
            len += 1;
        }
        if !self.encrypted_inputs_pointer.is_empty() {
            len += 1;
        }
        if !self.encrypted_outputs_pointer.is_empty() {
            len += 1;
        }
        if self.max_runtime_seconds != 0 {
            len += 1;
        }
        if !self.agreed_price.is_empty() {
            len += 1;
        }
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        if !self.scheduling_decision_id.is_empty() {
            len += 1;
        }
        if !self.status_message.is_empty() {
            len += 1;
        }
        if self.exit_code != 0 {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.queued_at.is_some() {
            len += 1;
        }
        if self.started_at.is_some() {
            len += 1;
        }
        if self.completed_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCJob", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.customer_address.is_empty() {
            struct_ser.serialize_field("customerAddress", &self.customer_address)?;
        }
        if !self.slurm_job_id.is_empty() {
            struct_ser.serialize_field("slurmJobId", &self.slurm_job_id)?;
        }
        if self.state != 0 {
            let v = JobState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.queue_name.is_empty() {
            struct_ser.serialize_field("queueName", &self.queue_name)?;
        }
        if let Some(v) = self.workload_spec.as_ref() {
            struct_ser.serialize_field("workloadSpec", v)?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        if !self.data_references.is_empty() {
            struct_ser.serialize_field("dataReferences", &self.data_references)?;
        }
        if !self.encrypted_inputs_pointer.is_empty() {
            struct_ser.serialize_field("encryptedInputsPointer", &self.encrypted_inputs_pointer)?;
        }
        if !self.encrypted_outputs_pointer.is_empty() {
            struct_ser.serialize_field("encryptedOutputsPointer", &self.encrypted_outputs_pointer)?;
        }
        if self.max_runtime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntimeSeconds", ToString::to_string(&self.max_runtime_seconds).as_str())?;
        }
        if !self.agreed_price.is_empty() {
            struct_ser.serialize_field("agreedPrice", &self.agreed_price)?;
        }
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        if !self.scheduling_decision_id.is_empty() {
            struct_ser.serialize_field("schedulingDecisionId", &self.scheduling_decision_id)?;
        }
        if !self.status_message.is_empty() {
            struct_ser.serialize_field("statusMessage", &self.status_message)?;
        }
        if self.exit_code != 0 {
            struct_ser.serialize_field("exitCode", &self.exit_code)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.queued_at.as_ref() {
            struct_ser.serialize_field("queuedAt", v)?;
        }
        if let Some(v) = self.started_at.as_ref() {
            struct_ser.serialize_field("startedAt", v)?;
        }
        if let Some(v) = self.completed_at.as_ref() {
            struct_ser.serialize_field("completedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcJob {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
            "offering_id",
            "offeringId",
            "cluster_id",
            "clusterId",
            "provider_address",
            "providerAddress",
            "customer_address",
            "customerAddress",
            "slurm_job_id",
            "slurmJobId",
            "state",
            "queue_name",
            "queueName",
            "workload_spec",
            "workloadSpec",
            "resources",
            "data_references",
            "dataReferences",
            "encrypted_inputs_pointer",
            "encryptedInputsPointer",
            "encrypted_outputs_pointer",
            "encryptedOutputsPointer",
            "max_runtime_seconds",
            "maxRuntimeSeconds",
            "agreed_price",
            "agreedPrice",
            "escrow_id",
            "escrowId",
            "scheduling_decision_id",
            "schedulingDecisionId",
            "status_message",
            "statusMessage",
            "exit_code",
            "exitCode",
            "created_at",
            "createdAt",
            "queued_at",
            "queuedAt",
            "started_at",
            "startedAt",
            "completed_at",
            "completedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
            OfferingId,
            ClusterId,
            ProviderAddress,
            CustomerAddress,
            SlurmJobId,
            State,
            QueueName,
            WorkloadSpec,
            Resources,
            DataReferences,
            EncryptedInputsPointer,
            EncryptedOutputsPointer,
            MaxRuntimeSeconds,
            AgreedPrice,
            EscrowId,
            SchedulingDecisionId,
            StatusMessage,
            ExitCode,
            CreatedAt,
            QueuedAt,
            StartedAt,
            CompletedAt,
            BlockHeight,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "customerAddress" | "customer_address" => Ok(GeneratedField::CustomerAddress),
                            "slurmJobId" | "slurm_job_id" => Ok(GeneratedField::SlurmJobId),
                            "state" => Ok(GeneratedField::State),
                            "queueName" | "queue_name" => Ok(GeneratedField::QueueName),
                            "workloadSpec" | "workload_spec" => Ok(GeneratedField::WorkloadSpec),
                            "resources" => Ok(GeneratedField::Resources),
                            "dataReferences" | "data_references" => Ok(GeneratedField::DataReferences),
                            "encryptedInputsPointer" | "encrypted_inputs_pointer" => Ok(GeneratedField::EncryptedInputsPointer),
                            "encryptedOutputsPointer" | "encrypted_outputs_pointer" => Ok(GeneratedField::EncryptedOutputsPointer),
                            "maxRuntimeSeconds" | "max_runtime_seconds" => Ok(GeneratedField::MaxRuntimeSeconds),
                            "agreedPrice" | "agreed_price" => Ok(GeneratedField::AgreedPrice),
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            "schedulingDecisionId" | "scheduling_decision_id" => Ok(GeneratedField::SchedulingDecisionId),
                            "statusMessage" | "status_message" => Ok(GeneratedField::StatusMessage),
                            "exitCode" | "exit_code" => Ok(GeneratedField::ExitCode),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "queuedAt" | "queued_at" => Ok(GeneratedField::QueuedAt),
                            "startedAt" | "started_at" => Ok(GeneratedField::StartedAt),
                            "completedAt" | "completed_at" => Ok(GeneratedField::CompletedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcJob;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCJob")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcJob, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                let mut offering_id__ = None;
                let mut cluster_id__ = None;
                let mut provider_address__ = None;
                let mut customer_address__ = None;
                let mut slurm_job_id__ = None;
                let mut state__ = None;
                let mut queue_name__ = None;
                let mut workload_spec__ = None;
                let mut resources__ = None;
                let mut data_references__ = None;
                let mut encrypted_inputs_pointer__ = None;
                let mut encrypted_outputs_pointer__ = None;
                let mut max_runtime_seconds__ = None;
                let mut agreed_price__ = None;
                let mut escrow_id__ = None;
                let mut scheduling_decision_id__ = None;
                let mut status_message__ = None;
                let mut exit_code__ = None;
                let mut created_at__ = None;
                let mut queued_at__ = None;
                let mut started_at__ = None;
                let mut completed_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CustomerAddress => {
                            if customer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customerAddress"));
                            }
                            customer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SlurmJobId => {
                            if slurm_job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slurmJobId"));
                            }
                            slurm_job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<JobState>()? as i32);
                        }
                        GeneratedField::QueueName => {
                            if queue_name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queueName"));
                            }
                            queue_name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::WorkloadSpec => {
                            if workload_spec__.is_some() {
                                return Err(serde::de::Error::duplicate_field("workloadSpec"));
                            }
                            workload_spec__ = map_.next_value()?;
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                        GeneratedField::DataReferences => {
                            if data_references__.is_some() {
                                return Err(serde::de::Error::duplicate_field("dataReferences"));
                            }
                            data_references__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptedInputsPointer => {
                            if encrypted_inputs_pointer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedInputsPointer"));
                            }
                            encrypted_inputs_pointer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptedOutputsPointer => {
                            if encrypted_outputs_pointer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedOutputsPointer"));
                            }
                            encrypted_outputs_pointer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxRuntimeSeconds => {
                            if max_runtime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntimeSeconds"));
                            }
                            max_runtime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AgreedPrice => {
                            if agreed_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("agreedPrice"));
                            }
                            agreed_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SchedulingDecisionId => {
                            if scheduling_decision_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("schedulingDecisionId"));
                            }
                            scheduling_decision_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StatusMessage => {
                            if status_message__.is_some() {
                                return Err(serde::de::Error::duplicate_field("statusMessage"));
                            }
                            status_message__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExitCode => {
                            if exit_code__.is_some() {
                                return Err(serde::de::Error::duplicate_field("exitCode"));
                            }
                            exit_code__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
                        }
                        GeneratedField::QueuedAt => {
                            if queued_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queuedAt"));
                            }
                            queued_at__ = map_.next_value()?;
                        }
                        GeneratedField::StartedAt => {
                            if started_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("startedAt"));
                            }
                            started_at__ = map_.next_value()?;
                        }
                        GeneratedField::CompletedAt => {
                            if completed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completedAt"));
                            }
                            completed_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(HpcJob {
                    job_id: job_id__.unwrap_or_default(),
                    offering_id: offering_id__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    customer_address: customer_address__.unwrap_or_default(),
                    slurm_job_id: slurm_job_id__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    queue_name: queue_name__.unwrap_or_default(),
                    workload_spec: workload_spec__,
                    resources: resources__,
                    data_references: data_references__.unwrap_or_default(),
                    encrypted_inputs_pointer: encrypted_inputs_pointer__.unwrap_or_default(),
                    encrypted_outputs_pointer: encrypted_outputs_pointer__.unwrap_or_default(),
                    max_runtime_seconds: max_runtime_seconds__.unwrap_or_default(),
                    agreed_price: agreed_price__.unwrap_or_default(),
                    escrow_id: escrow_id__.unwrap_or_default(),
                    scheduling_decision_id: scheduling_decision_id__.unwrap_or_default(),
                    status_message: status_message__.unwrap_or_default(),
                    exit_code: exit_code__.unwrap_or_default(),
                    created_at: created_at__,
                    queued_at: queued_at__,
                    started_at: started_at__,
                    completed_at: completed_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCJob", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcOffering {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.offering_id.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.queue_options.is_empty() {
            len += 1;
        }
        if self.pricing.is_some() {
            len += 1;
        }
        if self.required_identity_threshold != 0 {
            len += 1;
        }
        if self.max_runtime_seconds != 0 {
            len += 1;
        }
        if !self.preconfigured_workloads.is_empty() {
            len += 1;
        }
        if self.supports_custom_workloads {
            len += 1;
        }
        if self.active {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCOffering", len)?;
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.queue_options.is_empty() {
            struct_ser.serialize_field("queueOptions", &self.queue_options)?;
        }
        if let Some(v) = self.pricing.as_ref() {
            struct_ser.serialize_field("pricing", v)?;
        }
        if self.required_identity_threshold != 0 {
            struct_ser.serialize_field("requiredIdentityThreshold", &self.required_identity_threshold)?;
        }
        if self.max_runtime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntimeSeconds", ToString::to_string(&self.max_runtime_seconds).as_str())?;
        }
        if !self.preconfigured_workloads.is_empty() {
            struct_ser.serialize_field("preconfiguredWorkloads", &self.preconfigured_workloads)?;
        }
        if self.supports_custom_workloads {
            struct_ser.serialize_field("supportsCustomWorkloads", &self.supports_custom_workloads)?;
        }
        if self.active {
            struct_ser.serialize_field("active", &self.active)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcOffering {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offering_id",
            "offeringId",
            "cluster_id",
            "clusterId",
            "provider_address",
            "providerAddress",
            "name",
            "description",
            "queue_options",
            "queueOptions",
            "pricing",
            "required_identity_threshold",
            "requiredIdentityThreshold",
            "max_runtime_seconds",
            "maxRuntimeSeconds",
            "preconfigured_workloads",
            "preconfiguredWorkloads",
            "supports_custom_workloads",
            "supportsCustomWorkloads",
            "active",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            OfferingId,
            ClusterId,
            ProviderAddress,
            Name,
            Description,
            QueueOptions,
            Pricing,
            RequiredIdentityThreshold,
            MaxRuntimeSeconds,
            PreconfiguredWorkloads,
            SupportsCustomWorkloads,
            Active,
            CreatedAt,
            UpdatedAt,
            BlockHeight,
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
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "queueOptions" | "queue_options" => Ok(GeneratedField::QueueOptions),
                            "pricing" => Ok(GeneratedField::Pricing),
                            "requiredIdentityThreshold" | "required_identity_threshold" => Ok(GeneratedField::RequiredIdentityThreshold),
                            "maxRuntimeSeconds" | "max_runtime_seconds" => Ok(GeneratedField::MaxRuntimeSeconds),
                            "preconfiguredWorkloads" | "preconfigured_workloads" => Ok(GeneratedField::PreconfiguredWorkloads),
                            "supportsCustomWorkloads" | "supports_custom_workloads" => Ok(GeneratedField::SupportsCustomWorkloads),
                            "active" => Ok(GeneratedField::Active),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcOffering;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCOffering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcOffering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offering_id__ = None;
                let mut cluster_id__ = None;
                let mut provider_address__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut queue_options__ = None;
                let mut pricing__ = None;
                let mut required_identity_threshold__ = None;
                let mut max_runtime_seconds__ = None;
                let mut preconfigured_workloads__ = None;
                let mut supports_custom_workloads__ = None;
                let mut active__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::QueueOptions => {
                            if queue_options__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queueOptions"));
                            }
                            queue_options__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pricing => {
                            if pricing__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pricing"));
                            }
                            pricing__ = map_.next_value()?;
                        }
                        GeneratedField::RequiredIdentityThreshold => {
                            if required_identity_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredIdentityThreshold"));
                            }
                            required_identity_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRuntimeSeconds => {
                            if max_runtime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntimeSeconds"));
                            }
                            max_runtime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreconfiguredWorkloads => {
                            if preconfigured_workloads__.is_some() {
                                return Err(serde::de::Error::duplicate_field("preconfiguredWorkloads"));
                            }
                            preconfigured_workloads__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SupportsCustomWorkloads => {
                            if supports_custom_workloads__.is_some() {
                                return Err(serde::de::Error::duplicate_field("supportsCustomWorkloads"));
                            }
                            supports_custom_workloads__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(HpcOffering {
                    offering_id: offering_id__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    queue_options: queue_options__.unwrap_or_default(),
                    pricing: pricing__,
                    required_identity_threshold: required_identity_threshold__.unwrap_or_default(),
                    max_runtime_seconds: max_runtime_seconds__.unwrap_or_default(),
                    preconfigured_workloads: preconfigured_workloads__.unwrap_or_default(),
                    supports_custom_workloads: supports_custom_workloads__.unwrap_or_default(),
                    active: active__.unwrap_or_default(),
                    created_at: created_at__,
                    updated_at: updated_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCOffering", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcPricing {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.base_node_hour_price.is_empty() {
            len += 1;
        }
        if !self.cpu_core_hour_price.is_empty() {
            len += 1;
        }
        if !self.gpu_hour_price.is_empty() {
            len += 1;
        }
        if !self.memory_gb_hour_price.is_empty() {
            len += 1;
        }
        if !self.storage_gb_price.is_empty() {
            len += 1;
        }
        if !self.network_gb_price.is_empty() {
            len += 1;
        }
        if !self.currency.is_empty() {
            len += 1;
        }
        if !self.minimum_charge.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCPricing", len)?;
        if !self.base_node_hour_price.is_empty() {
            struct_ser.serialize_field("baseNodeHourPrice", &self.base_node_hour_price)?;
        }
        if !self.cpu_core_hour_price.is_empty() {
            struct_ser.serialize_field("cpuCoreHourPrice", &self.cpu_core_hour_price)?;
        }
        if !self.gpu_hour_price.is_empty() {
            struct_ser.serialize_field("gpuHourPrice", &self.gpu_hour_price)?;
        }
        if !self.memory_gb_hour_price.is_empty() {
            struct_ser.serialize_field("memoryGbHourPrice", &self.memory_gb_hour_price)?;
        }
        if !self.storage_gb_price.is_empty() {
            struct_ser.serialize_field("storageGbPrice", &self.storage_gb_price)?;
        }
        if !self.network_gb_price.is_empty() {
            struct_ser.serialize_field("networkGbPrice", &self.network_gb_price)?;
        }
        if !self.currency.is_empty() {
            struct_ser.serialize_field("currency", &self.currency)?;
        }
        if !self.minimum_charge.is_empty() {
            struct_ser.serialize_field("minimumCharge", &self.minimum_charge)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcPricing {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "base_node_hour_price",
            "baseNodeHourPrice",
            "cpu_core_hour_price",
            "cpuCoreHourPrice",
            "gpu_hour_price",
            "gpuHourPrice",
            "memory_gb_hour_price",
            "memoryGbHourPrice",
            "storage_gb_price",
            "storageGbPrice",
            "network_gb_price",
            "networkGbPrice",
            "currency",
            "minimum_charge",
            "minimumCharge",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            BaseNodeHourPrice,
            CpuCoreHourPrice,
            GpuHourPrice,
            MemoryGbHourPrice,
            StorageGbPrice,
            NetworkGbPrice,
            Currency,
            MinimumCharge,
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
                            "baseNodeHourPrice" | "base_node_hour_price" => Ok(GeneratedField::BaseNodeHourPrice),
                            "cpuCoreHourPrice" | "cpu_core_hour_price" => Ok(GeneratedField::CpuCoreHourPrice),
                            "gpuHourPrice" | "gpu_hour_price" => Ok(GeneratedField::GpuHourPrice),
                            "memoryGbHourPrice" | "memory_gb_hour_price" => Ok(GeneratedField::MemoryGbHourPrice),
                            "storageGbPrice" | "storage_gb_price" => Ok(GeneratedField::StorageGbPrice),
                            "networkGbPrice" | "network_gb_price" => Ok(GeneratedField::NetworkGbPrice),
                            "currency" => Ok(GeneratedField::Currency),
                            "minimumCharge" | "minimum_charge" => Ok(GeneratedField::MinimumCharge),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcPricing;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCPricing")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcPricing, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut base_node_hour_price__ = None;
                let mut cpu_core_hour_price__ = None;
                let mut gpu_hour_price__ = None;
                let mut memory_gb_hour_price__ = None;
                let mut storage_gb_price__ = None;
                let mut network_gb_price__ = None;
                let mut currency__ = None;
                let mut minimum_charge__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::BaseNodeHourPrice => {
                            if base_node_hour_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("baseNodeHourPrice"));
                            }
                            base_node_hour_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CpuCoreHourPrice => {
                            if cpu_core_hour_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cpuCoreHourPrice"));
                            }
                            cpu_core_hour_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GpuHourPrice => {
                            if gpu_hour_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpuHourPrice"));
                            }
                            gpu_hour_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MemoryGbHourPrice => {
                            if memory_gb_hour_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("memoryGbHourPrice"));
                            }
                            memory_gb_hour_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StorageGbPrice => {
                            if storage_gb_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storageGbPrice"));
                            }
                            storage_gb_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NetworkGbPrice => {
                            if network_gb_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("networkGbPrice"));
                            }
                            network_gb_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Currency => {
                            if currency__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currency"));
                            }
                            currency__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinimumCharge => {
                            if minimum_charge__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minimumCharge"));
                            }
                            minimum_charge__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(HpcPricing {
                    base_node_hour_price: base_node_hour_price__.unwrap_or_default(),
                    cpu_core_hour_price: cpu_core_hour_price__.unwrap_or_default(),
                    gpu_hour_price: gpu_hour_price__.unwrap_or_default(),
                    memory_gb_hour_price: memory_gb_hour_price__.unwrap_or_default(),
                    storage_gb_price: storage_gb_price__.unwrap_or_default(),
                    network_gb_price: network_gb_price__.unwrap_or_default(),
                    currency: currency__.unwrap_or_default(),
                    minimum_charge: minimum_charge__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCPricing", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcRewardRecipient {
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
        if !self.amount.is_empty() {
            len += 1;
        }
        if !self.recipient_type.is_empty() {
            len += 1;
        }
        if !self.node_id.is_empty() {
            len += 1;
        }
        if !self.contribution_weight.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCRewardRecipient", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        if !self.recipient_type.is_empty() {
            struct_ser.serialize_field("recipientType", &self.recipient_type)?;
        }
        if !self.node_id.is_empty() {
            struct_ser.serialize_field("nodeId", &self.node_id)?;
        }
        if !self.contribution_weight.is_empty() {
            struct_ser.serialize_field("contributionWeight", &self.contribution_weight)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcRewardRecipient {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "amount",
            "recipient_type",
            "recipientType",
            "node_id",
            "nodeId",
            "contribution_weight",
            "contributionWeight",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Amount,
            RecipientType,
            NodeId,
            ContributionWeight,
            Reason,
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
                            "amount" => Ok(GeneratedField::Amount),
                            "recipientType" | "recipient_type" => Ok(GeneratedField::RecipientType),
                            "nodeId" | "node_id" => Ok(GeneratedField::NodeId),
                            "contributionWeight" | "contribution_weight" => Ok(GeneratedField::ContributionWeight),
                            "reason" => Ok(GeneratedField::Reason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcRewardRecipient;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCRewardRecipient")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcRewardRecipient, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut amount__ = None;
                let mut recipient_type__ = None;
                let mut node_id__ = None;
                let mut contribution_weight__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecipientType => {
                            if recipient_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientType"));
                            }
                            recipient_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NodeId => {
                            if node_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeId"));
                            }
                            node_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ContributionWeight => {
                            if contribution_weight__.is_some() {
                                return Err(serde::de::Error::duplicate_field("contributionWeight"));
                            }
                            contribution_weight__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(HpcRewardRecipient {
                    address: address__.unwrap_or_default(),
                    amount: amount__.unwrap_or_default(),
                    recipient_type: recipient_type__.unwrap_or_default(),
                    node_id: node_id__.unwrap_or_default(),
                    contribution_weight: contribution_weight__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCRewardRecipient", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcRewardRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reward_id.is_empty() {
            len += 1;
        }
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if self.source != 0 {
            len += 1;
        }
        if !self.total_reward.is_empty() {
            len += 1;
        }
        if !self.recipients.is_empty() {
            len += 1;
        }
        if !self.referenced_usage_records.is_empty() {
            len += 1;
        }
        if self.job_completion_status != 0 {
            len += 1;
        }
        if !self.formula_version.is_empty() {
            len += 1;
        }
        if self.calculation_details.is_some() {
            len += 1;
        }
        if self.disputed {
            len += 1;
        }
        if !self.dispute_id.is_empty() {
            len += 1;
        }
        if self.issued_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCRewardRecord", len)?;
        if !self.reward_id.is_empty() {
            struct_ser.serialize_field("rewardId", &self.reward_id)?;
        }
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if self.source != 0 {
            let v = HpcRewardSource::try_from(self.source)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.source)))?;
            struct_ser.serialize_field("source", &v)?;
        }
        if !self.total_reward.is_empty() {
            struct_ser.serialize_field("totalReward", &self.total_reward)?;
        }
        if !self.recipients.is_empty() {
            struct_ser.serialize_field("recipients", &self.recipients)?;
        }
        if !self.referenced_usage_records.is_empty() {
            struct_ser.serialize_field("referencedUsageRecords", &self.referenced_usage_records)?;
        }
        if self.job_completion_status != 0 {
            let v = JobState::try_from(self.job_completion_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.job_completion_status)))?;
            struct_ser.serialize_field("jobCompletionStatus", &v)?;
        }
        if !self.formula_version.is_empty() {
            struct_ser.serialize_field("formulaVersion", &self.formula_version)?;
        }
        if let Some(v) = self.calculation_details.as_ref() {
            struct_ser.serialize_field("calculationDetails", v)?;
        }
        if self.disputed {
            struct_ser.serialize_field("disputed", &self.disputed)?;
        }
        if !self.dispute_id.is_empty() {
            struct_ser.serialize_field("disputeId", &self.dispute_id)?;
        }
        if let Some(v) = self.issued_at.as_ref() {
            struct_ser.serialize_field("issuedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcRewardRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reward_id",
            "rewardId",
            "job_id",
            "jobId",
            "cluster_id",
            "clusterId",
            "source",
            "total_reward",
            "totalReward",
            "recipients",
            "referenced_usage_records",
            "referencedUsageRecords",
            "job_completion_status",
            "jobCompletionStatus",
            "formula_version",
            "formulaVersion",
            "calculation_details",
            "calculationDetails",
            "disputed",
            "dispute_id",
            "disputeId",
            "issued_at",
            "issuedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RewardId,
            JobId,
            ClusterId,
            Source,
            TotalReward,
            Recipients,
            ReferencedUsageRecords,
            JobCompletionStatus,
            FormulaVersion,
            CalculationDetails,
            Disputed,
            DisputeId,
            IssuedAt,
            BlockHeight,
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
                            "rewardId" | "reward_id" => Ok(GeneratedField::RewardId),
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "source" => Ok(GeneratedField::Source),
                            "totalReward" | "total_reward" => Ok(GeneratedField::TotalReward),
                            "recipients" => Ok(GeneratedField::Recipients),
                            "referencedUsageRecords" | "referenced_usage_records" => Ok(GeneratedField::ReferencedUsageRecords),
                            "jobCompletionStatus" | "job_completion_status" => Ok(GeneratedField::JobCompletionStatus),
                            "formulaVersion" | "formula_version" => Ok(GeneratedField::FormulaVersion),
                            "calculationDetails" | "calculation_details" => Ok(GeneratedField::CalculationDetails),
                            "disputed" => Ok(GeneratedField::Disputed),
                            "disputeId" | "dispute_id" => Ok(GeneratedField::DisputeId),
                            "issuedAt" | "issued_at" => Ok(GeneratedField::IssuedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcRewardRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCRewardRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcRewardRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reward_id__ = None;
                let mut job_id__ = None;
                let mut cluster_id__ = None;
                let mut source__ = None;
                let mut total_reward__ = None;
                let mut recipients__ = None;
                let mut referenced_usage_records__ = None;
                let mut job_completion_status__ = None;
                let mut formula_version__ = None;
                let mut calculation_details__ = None;
                let mut disputed__ = None;
                let mut dispute_id__ = None;
                let mut issued_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RewardId => {
                            if reward_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardId"));
                            }
                            reward_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value::<HpcRewardSource>()? as i32);
                        }
                        GeneratedField::TotalReward => {
                            if total_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalReward"));
                            }
                            total_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Recipients => {
                            if recipients__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipients"));
                            }
                            recipients__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReferencedUsageRecords => {
                            if referenced_usage_records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("referencedUsageRecords"));
                            }
                            referenced_usage_records__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobCompletionStatus => {
                            if job_completion_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobCompletionStatus"));
                            }
                            job_completion_status__ = Some(map_.next_value::<JobState>()? as i32);
                        }
                        GeneratedField::FormulaVersion => {
                            if formula_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("formulaVersion"));
                            }
                            formula_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CalculationDetails => {
                            if calculation_details__.is_some() {
                                return Err(serde::de::Error::duplicate_field("calculationDetails"));
                            }
                            calculation_details__ = map_.next_value()?;
                        }
                        GeneratedField::Disputed => {
                            if disputed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputed"));
                            }
                            disputed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DisputeId => {
                            if dispute_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeId"));
                            }
                            dispute_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IssuedAt => {
                            if issued_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("issuedAt"));
                            }
                            issued_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(HpcRewardRecord {
                    reward_id: reward_id__.unwrap_or_default(),
                    job_id: job_id__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    source: source__.unwrap_or_default(),
                    total_reward: total_reward__.unwrap_or_default(),
                    recipients: recipients__.unwrap_or_default(),
                    referenced_usage_records: referenced_usage_records__.unwrap_or_default(),
                    job_completion_status: job_completion_status__.unwrap_or_default(),
                    formula_version: formula_version__.unwrap_or_default(),
                    calculation_details: calculation_details__,
                    disputed: disputed__.unwrap_or_default(),
                    dispute_id: dispute_id__.unwrap_or_default(),
                    issued_at: issued_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCRewardRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for HpcRewardSource {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "HPC_REWARD_SOURCE_UNSPECIFIED",
            Self::JobCompletion => "HPC_REWARD_SOURCE_JOB_COMPLETION",
            Self::Usage => "HPC_REWARD_SOURCE_USAGE",
            Self::Bonus => "HPC_REWARD_SOURCE_BONUS",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for HpcRewardSource {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "HPC_REWARD_SOURCE_UNSPECIFIED",
            "HPC_REWARD_SOURCE_JOB_COMPLETION",
            "HPC_REWARD_SOURCE_USAGE",
            "HPC_REWARD_SOURCE_BONUS",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcRewardSource;

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
                    "HPC_REWARD_SOURCE_UNSPECIFIED" => Ok(HpcRewardSource::Unspecified),
                    "HPC_REWARD_SOURCE_JOB_COMPLETION" => Ok(HpcRewardSource::JobCompletion),
                    "HPC_REWARD_SOURCE_USAGE" => Ok(HpcRewardSource::Usage),
                    "HPC_REWARD_SOURCE_BONUS" => Ok(HpcRewardSource::Bonus),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for HpcUsageMetrics {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.wall_clock_seconds != 0 {
            len += 1;
        }
        if self.cpu_core_seconds != 0 {
            len += 1;
        }
        if self.memory_gb_seconds != 0 {
            len += 1;
        }
        if self.gpu_seconds != 0 {
            len += 1;
        }
        if self.storage_gb_hours != 0 {
            len += 1;
        }
        if self.network_bytes_in != 0 {
            len += 1;
        }
        if self.network_bytes_out != 0 {
            len += 1;
        }
        if self.node_hours != 0 {
            len += 1;
        }
        if self.nodes_used != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.HPCUsageMetrics", len)?;
        if self.wall_clock_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("wallClockSeconds", ToString::to_string(&self.wall_clock_seconds).as_str())?;
        }
        if self.cpu_core_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("cpuCoreSeconds", ToString::to_string(&self.cpu_core_seconds).as_str())?;
        }
        if self.memory_gb_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("memoryGbSeconds", ToString::to_string(&self.memory_gb_seconds).as_str())?;
        }
        if self.gpu_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("gpuSeconds", ToString::to_string(&self.gpu_seconds).as_str())?;
        }
        if self.storage_gb_hours != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("storageGbHours", ToString::to_string(&self.storage_gb_hours).as_str())?;
        }
        if self.network_bytes_in != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("networkBytesIn", ToString::to_string(&self.network_bytes_in).as_str())?;
        }
        if self.network_bytes_out != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("networkBytesOut", ToString::to_string(&self.network_bytes_out).as_str())?;
        }
        if self.node_hours != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nodeHours", ToString::to_string(&self.node_hours).as_str())?;
        }
        if self.nodes_used != 0 {
            struct_ser.serialize_field("nodesUsed", &self.nodes_used)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for HpcUsageMetrics {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "wall_clock_seconds",
            "wallClockSeconds",
            "cpu_core_seconds",
            "cpuCoreSeconds",
            "memory_gb_seconds",
            "memoryGbSeconds",
            "gpu_seconds",
            "gpuSeconds",
            "storage_gb_hours",
            "storageGbHours",
            "network_bytes_in",
            "networkBytesIn",
            "network_bytes_out",
            "networkBytesOut",
            "node_hours",
            "nodeHours",
            "nodes_used",
            "nodesUsed",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            WallClockSeconds,
            CpuCoreSeconds,
            MemoryGbSeconds,
            GpuSeconds,
            StorageGbHours,
            NetworkBytesIn,
            NetworkBytesOut,
            NodeHours,
            NodesUsed,
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
                            "wallClockSeconds" | "wall_clock_seconds" => Ok(GeneratedField::WallClockSeconds),
                            "cpuCoreSeconds" | "cpu_core_seconds" => Ok(GeneratedField::CpuCoreSeconds),
                            "memoryGbSeconds" | "memory_gb_seconds" => Ok(GeneratedField::MemoryGbSeconds),
                            "gpuSeconds" | "gpu_seconds" => Ok(GeneratedField::GpuSeconds),
                            "storageGbHours" | "storage_gb_hours" => Ok(GeneratedField::StorageGbHours),
                            "networkBytesIn" | "network_bytes_in" => Ok(GeneratedField::NetworkBytesIn),
                            "networkBytesOut" | "network_bytes_out" => Ok(GeneratedField::NetworkBytesOut),
                            "nodeHours" | "node_hours" => Ok(GeneratedField::NodeHours),
                            "nodesUsed" | "nodes_used" => Ok(GeneratedField::NodesUsed),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = HpcUsageMetrics;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.HPCUsageMetrics")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<HpcUsageMetrics, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut wall_clock_seconds__ = None;
                let mut cpu_core_seconds__ = None;
                let mut memory_gb_seconds__ = None;
                let mut gpu_seconds__ = None;
                let mut storage_gb_hours__ = None;
                let mut network_bytes_in__ = None;
                let mut network_bytes_out__ = None;
                let mut node_hours__ = None;
                let mut nodes_used__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::WallClockSeconds => {
                            if wall_clock_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("wallClockSeconds"));
                            }
                            wall_clock_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CpuCoreSeconds => {
                            if cpu_core_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cpuCoreSeconds"));
                            }
                            cpu_core_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MemoryGbSeconds => {
                            if memory_gb_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("memoryGbSeconds"));
                            }
                            memory_gb_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GpuSeconds => {
                            if gpu_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpuSeconds"));
                            }
                            gpu_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::StorageGbHours => {
                            if storage_gb_hours__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storageGbHours"));
                            }
                            storage_gb_hours__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NetworkBytesIn => {
                            if network_bytes_in__.is_some() {
                                return Err(serde::de::Error::duplicate_field("networkBytesIn"));
                            }
                            network_bytes_in__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NetworkBytesOut => {
                            if network_bytes_out__.is_some() {
                                return Err(serde::de::Error::duplicate_field("networkBytesOut"));
                            }
                            network_bytes_out__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NodeHours => {
                            if node_hours__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeHours"));
                            }
                            node_hours__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NodesUsed => {
                            if nodes_used__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodesUsed"));
                            }
                            nodes_used__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(HpcUsageMetrics {
                    wall_clock_seconds: wall_clock_seconds__.unwrap_or_default(),
                    cpu_core_seconds: cpu_core_seconds__.unwrap_or_default(),
                    memory_gb_seconds: memory_gb_seconds__.unwrap_or_default(),
                    gpu_seconds: gpu_seconds__.unwrap_or_default(),
                    storage_gb_hours: storage_gb_hours__.unwrap_or_default(),
                    network_bytes_in: network_bytes_in__.unwrap_or_default(),
                    network_bytes_out: network_bytes_out__.unwrap_or_default(),
                    node_hours: node_hours__.unwrap_or_default(),
                    nodes_used: nodes_used__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.HPCUsageMetrics", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for JobAccounting {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.customer_address.is_empty() {
            len += 1;
        }
        if self.usage_metrics.is_some() {
            len += 1;
        }
        if !self.total_cost.is_empty() {
            len += 1;
        }
        if !self.provider_reward.is_empty() {
            len += 1;
        }
        if !self.node_rewards.is_empty() {
            len += 1;
        }
        if !self.platform_fee.is_empty() {
            len += 1;
        }
        if !self.settlement_status.is_empty() {
            len += 1;
        }
        if !self.settlement_id.is_empty() {
            len += 1;
        }
        if !self.signed_usage_record_ids.is_empty() {
            len += 1;
        }
        if self.job_completion_status != 0 {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.finalized_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.JobAccounting", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.customer_address.is_empty() {
            struct_ser.serialize_field("customerAddress", &self.customer_address)?;
        }
        if let Some(v) = self.usage_metrics.as_ref() {
            struct_ser.serialize_field("usageMetrics", v)?;
        }
        if !self.total_cost.is_empty() {
            struct_ser.serialize_field("totalCost", &self.total_cost)?;
        }
        if !self.provider_reward.is_empty() {
            struct_ser.serialize_field("providerReward", &self.provider_reward)?;
        }
        if !self.node_rewards.is_empty() {
            struct_ser.serialize_field("nodeRewards", &self.node_rewards)?;
        }
        if !self.platform_fee.is_empty() {
            struct_ser.serialize_field("platformFee", &self.platform_fee)?;
        }
        if !self.settlement_status.is_empty() {
            struct_ser.serialize_field("settlementStatus", &self.settlement_status)?;
        }
        if !self.settlement_id.is_empty() {
            struct_ser.serialize_field("settlementId", &self.settlement_id)?;
        }
        if !self.signed_usage_record_ids.is_empty() {
            struct_ser.serialize_field("signedUsageRecordIds", &self.signed_usage_record_ids)?;
        }
        if self.job_completion_status != 0 {
            let v = JobState::try_from(self.job_completion_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.job_completion_status)))?;
            struct_ser.serialize_field("jobCompletionStatus", &v)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.finalized_at.as_ref() {
            struct_ser.serialize_field("finalizedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for JobAccounting {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
            "cluster_id",
            "clusterId",
            "provider_address",
            "providerAddress",
            "customer_address",
            "customerAddress",
            "usage_metrics",
            "usageMetrics",
            "total_cost",
            "totalCost",
            "provider_reward",
            "providerReward",
            "node_rewards",
            "nodeRewards",
            "platform_fee",
            "platformFee",
            "settlement_status",
            "settlementStatus",
            "settlement_id",
            "settlementId",
            "signed_usage_record_ids",
            "signedUsageRecordIds",
            "job_completion_status",
            "jobCompletionStatus",
            "created_at",
            "createdAt",
            "finalized_at",
            "finalizedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
            ClusterId,
            ProviderAddress,
            CustomerAddress,
            UsageMetrics,
            TotalCost,
            ProviderReward,
            NodeRewards,
            PlatformFee,
            SettlementStatus,
            SettlementId,
            SignedUsageRecordIds,
            JobCompletionStatus,
            CreatedAt,
            FinalizedAt,
            BlockHeight,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "customerAddress" | "customer_address" => Ok(GeneratedField::CustomerAddress),
                            "usageMetrics" | "usage_metrics" => Ok(GeneratedField::UsageMetrics),
                            "totalCost" | "total_cost" => Ok(GeneratedField::TotalCost),
                            "providerReward" | "provider_reward" => Ok(GeneratedField::ProviderReward),
                            "nodeRewards" | "node_rewards" => Ok(GeneratedField::NodeRewards),
                            "platformFee" | "platform_fee" => Ok(GeneratedField::PlatformFee),
                            "settlementStatus" | "settlement_status" => Ok(GeneratedField::SettlementStatus),
                            "settlementId" | "settlement_id" => Ok(GeneratedField::SettlementId),
                            "signedUsageRecordIds" | "signed_usage_record_ids" => Ok(GeneratedField::SignedUsageRecordIds),
                            "jobCompletionStatus" | "job_completion_status" => Ok(GeneratedField::JobCompletionStatus),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "finalizedAt" | "finalized_at" => Ok(GeneratedField::FinalizedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = JobAccounting;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.JobAccounting")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<JobAccounting, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                let mut cluster_id__ = None;
                let mut provider_address__ = None;
                let mut customer_address__ = None;
                let mut usage_metrics__ = None;
                let mut total_cost__ = None;
                let mut provider_reward__ = None;
                let mut node_rewards__ = None;
                let mut platform_fee__ = None;
                let mut settlement_status__ = None;
                let mut settlement_id__ = None;
                let mut signed_usage_record_ids__ = None;
                let mut job_completion_status__ = None;
                let mut created_at__ = None;
                let mut finalized_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CustomerAddress => {
                            if customer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customerAddress"));
                            }
                            customer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UsageMetrics => {
                            if usage_metrics__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageMetrics"));
                            }
                            usage_metrics__ = map_.next_value()?;
                        }
                        GeneratedField::TotalCost => {
                            if total_cost__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalCost"));
                            }
                            total_cost__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderReward => {
                            if provider_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerReward"));
                            }
                            provider_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NodeRewards => {
                            if node_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeRewards"));
                            }
                            node_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PlatformFee => {
                            if platform_fee__.is_some() {
                                return Err(serde::de::Error::duplicate_field("platformFee"));
                            }
                            platform_fee__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SettlementStatus => {
                            if settlement_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("settlementStatus"));
                            }
                            settlement_status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SettlementId => {
                            if settlement_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("settlementId"));
                            }
                            settlement_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SignedUsageRecordIds => {
                            if signed_usage_record_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signedUsageRecordIds"));
                            }
                            signed_usage_record_ids__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobCompletionStatus => {
                            if job_completion_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobCompletionStatus"));
                            }
                            job_completion_status__ = Some(map_.next_value::<JobState>()? as i32);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
                        }
                        GeneratedField::FinalizedAt => {
                            if finalized_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("finalizedAt"));
                            }
                            finalized_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(JobAccounting {
                    job_id: job_id__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    customer_address: customer_address__.unwrap_or_default(),
                    usage_metrics: usage_metrics__,
                    total_cost: total_cost__.unwrap_or_default(),
                    provider_reward: provider_reward__.unwrap_or_default(),
                    node_rewards: node_rewards__.unwrap_or_default(),
                    platform_fee: platform_fee__.unwrap_or_default(),
                    settlement_status: settlement_status__.unwrap_or_default(),
                    settlement_id: settlement_id__.unwrap_or_default(),
                    signed_usage_record_ids: signed_usage_record_ids__.unwrap_or_default(),
                    job_completion_status: job_completion_status__.unwrap_or_default(),
                    created_at: created_at__,
                    finalized_at: finalized_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.JobAccounting", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for JobResources {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.nodes != 0 {
            len += 1;
        }
        if self.cpu_cores_per_node != 0 {
            len += 1;
        }
        if self.memory_gb_per_node != 0 {
            len += 1;
        }
        if self.gpus_per_node != 0 {
            len += 1;
        }
        if self.storage_gb != 0 {
            len += 1;
        }
        if !self.gpu_type.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.JobResources", len)?;
        if self.nodes != 0 {
            struct_ser.serialize_field("nodes", &self.nodes)?;
        }
        if self.cpu_cores_per_node != 0 {
            struct_ser.serialize_field("cpuCoresPerNode", &self.cpu_cores_per_node)?;
        }
        if self.memory_gb_per_node != 0 {
            struct_ser.serialize_field("memoryGbPerNode", &self.memory_gb_per_node)?;
        }
        if self.gpus_per_node != 0 {
            struct_ser.serialize_field("gpusPerNode", &self.gpus_per_node)?;
        }
        if self.storage_gb != 0 {
            struct_ser.serialize_field("storageGb", &self.storage_gb)?;
        }
        if !self.gpu_type.is_empty() {
            struct_ser.serialize_field("gpuType", &self.gpu_type)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for JobResources {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "nodes",
            "cpu_cores_per_node",
            "cpuCoresPerNode",
            "memory_gb_per_node",
            "memoryGbPerNode",
            "gpus_per_node",
            "gpusPerNode",
            "storage_gb",
            "storageGb",
            "gpu_type",
            "gpuType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Nodes,
            CpuCoresPerNode,
            MemoryGbPerNode,
            GpusPerNode,
            StorageGb,
            GpuType,
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
                            "cpuCoresPerNode" | "cpu_cores_per_node" => Ok(GeneratedField::CpuCoresPerNode),
                            "memoryGbPerNode" | "memory_gb_per_node" => Ok(GeneratedField::MemoryGbPerNode),
                            "gpusPerNode" | "gpus_per_node" => Ok(GeneratedField::GpusPerNode),
                            "storageGb" | "storage_gb" => Ok(GeneratedField::StorageGb),
                            "gpuType" | "gpu_type" => Ok(GeneratedField::GpuType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = JobResources;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.JobResources")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<JobResources, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut nodes__ = None;
                let mut cpu_cores_per_node__ = None;
                let mut memory_gb_per_node__ = None;
                let mut gpus_per_node__ = None;
                let mut storage_gb__ = None;
                let mut gpu_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Nodes => {
                            if nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodes"));
                            }
                            nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CpuCoresPerNode => {
                            if cpu_cores_per_node__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cpuCoresPerNode"));
                            }
                            cpu_cores_per_node__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MemoryGbPerNode => {
                            if memory_gb_per_node__.is_some() {
                                return Err(serde::de::Error::duplicate_field("memoryGbPerNode"));
                            }
                            memory_gb_per_node__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GpusPerNode => {
                            if gpus_per_node__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpusPerNode"));
                            }
                            gpus_per_node__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::StorageGb => {
                            if storage_gb__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storageGb"));
                            }
                            storage_gb__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GpuType => {
                            if gpu_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpuType"));
                            }
                            gpu_type__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(JobResources {
                    nodes: nodes__.unwrap_or_default(),
                    cpu_cores_per_node: cpu_cores_per_node__.unwrap_or_default(),
                    memory_gb_per_node: memory_gb_per_node__.unwrap_or_default(),
                    gpus_per_node: gpus_per_node__.unwrap_or_default(),
                    storage_gb: storage_gb__.unwrap_or_default(),
                    gpu_type: gpu_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.JobResources", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for JobState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "JOB_STATE_UNSPECIFIED",
            Self::Pending => "JOB_STATE_PENDING",
            Self::Queued => "JOB_STATE_QUEUED",
            Self::Running => "JOB_STATE_RUNNING",
            Self::Completed => "JOB_STATE_COMPLETED",
            Self::Failed => "JOB_STATE_FAILED",
            Self::Cancelled => "JOB_STATE_CANCELLED",
            Self::Timeout => "JOB_STATE_TIMEOUT",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for JobState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "JOB_STATE_UNSPECIFIED",
            "JOB_STATE_PENDING",
            "JOB_STATE_QUEUED",
            "JOB_STATE_RUNNING",
            "JOB_STATE_COMPLETED",
            "JOB_STATE_FAILED",
            "JOB_STATE_CANCELLED",
            "JOB_STATE_TIMEOUT",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = JobState;

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
                    "JOB_STATE_UNSPECIFIED" => Ok(JobState::Unspecified),
                    "JOB_STATE_PENDING" => Ok(JobState::Pending),
                    "JOB_STATE_QUEUED" => Ok(JobState::Queued),
                    "JOB_STATE_RUNNING" => Ok(JobState::Running),
                    "JOB_STATE_COMPLETED" => Ok(JobState::Completed),
                    "JOB_STATE_FAILED" => Ok(JobState::Failed),
                    "JOB_STATE_CANCELLED" => Ok(JobState::Cancelled),
                    "JOB_STATE_TIMEOUT" => Ok(JobState::Timeout),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for JobWorkloadSpec {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.container_image.is_empty() {
            len += 1;
        }
        if !self.command.is_empty() {
            len += 1;
        }
        if !self.arguments.is_empty() {
            len += 1;
        }
        if !self.environment.is_empty() {
            len += 1;
        }
        if !self.working_directory.is_empty() {
            len += 1;
        }
        if !self.preconfigured_workload_id.is_empty() {
            len += 1;
        }
        if self.is_preconfigured {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.JobWorkloadSpec", len)?;
        if !self.container_image.is_empty() {
            struct_ser.serialize_field("containerImage", &self.container_image)?;
        }
        if !self.command.is_empty() {
            struct_ser.serialize_field("command", &self.command)?;
        }
        if !self.arguments.is_empty() {
            struct_ser.serialize_field("arguments", &self.arguments)?;
        }
        if !self.environment.is_empty() {
            struct_ser.serialize_field("environment", &self.environment)?;
        }
        if !self.working_directory.is_empty() {
            struct_ser.serialize_field("workingDirectory", &self.working_directory)?;
        }
        if !self.preconfigured_workload_id.is_empty() {
            struct_ser.serialize_field("preconfiguredWorkloadId", &self.preconfigured_workload_id)?;
        }
        if self.is_preconfigured {
            struct_ser.serialize_field("isPreconfigured", &self.is_preconfigured)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for JobWorkloadSpec {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "container_image",
            "containerImage",
            "command",
            "arguments",
            "environment",
            "working_directory",
            "workingDirectory",
            "preconfigured_workload_id",
            "preconfiguredWorkloadId",
            "is_preconfigured",
            "isPreconfigured",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ContainerImage,
            Command,
            Arguments,
            Environment,
            WorkingDirectory,
            PreconfiguredWorkloadId,
            IsPreconfigured,
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
                            "containerImage" | "container_image" => Ok(GeneratedField::ContainerImage),
                            "command" => Ok(GeneratedField::Command),
                            "arguments" => Ok(GeneratedField::Arguments),
                            "environment" => Ok(GeneratedField::Environment),
                            "workingDirectory" | "working_directory" => Ok(GeneratedField::WorkingDirectory),
                            "preconfiguredWorkloadId" | "preconfigured_workload_id" => Ok(GeneratedField::PreconfiguredWorkloadId),
                            "isPreconfigured" | "is_preconfigured" => Ok(GeneratedField::IsPreconfigured),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = JobWorkloadSpec;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.JobWorkloadSpec")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<JobWorkloadSpec, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut container_image__ = None;
                let mut command__ = None;
                let mut arguments__ = None;
                let mut environment__ = None;
                let mut working_directory__ = None;
                let mut preconfigured_workload_id__ = None;
                let mut is_preconfigured__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ContainerImage => {
                            if container_image__.is_some() {
                                return Err(serde::de::Error::duplicate_field("containerImage"));
                            }
                            container_image__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Command => {
                            if command__.is_some() {
                                return Err(serde::de::Error::duplicate_field("command"));
                            }
                            command__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Arguments => {
                            if arguments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("arguments"));
                            }
                            arguments__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Environment => {
                            if environment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("environment"));
                            }
                            environment__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                        GeneratedField::WorkingDirectory => {
                            if working_directory__.is_some() {
                                return Err(serde::de::Error::duplicate_field("workingDirectory"));
                            }
                            working_directory__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PreconfiguredWorkloadId => {
                            if preconfigured_workload_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("preconfiguredWorkloadId"));
                            }
                            preconfigured_workload_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsPreconfigured => {
                            if is_preconfigured__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isPreconfigured"));
                            }
                            is_preconfigured__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(JobWorkloadSpec {
                    container_image: container_image__.unwrap_or_default(),
                    command: command__.unwrap_or_default(),
                    arguments: arguments__.unwrap_or_default(),
                    environment: environment__.unwrap_or_default(),
                    working_directory: working_directory__.unwrap_or_default(),
                    preconfigured_workload_id: preconfigured_workload_id__.unwrap_or_default(),
                    is_preconfigured: is_preconfigured__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.JobWorkloadSpec", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LatencyMeasurement {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.target_node_id.is_empty() {
            len += 1;
        }
        if self.latency_ms != 0 {
            len += 1;
        }
        if self.measured_at.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.LatencyMeasurement", len)?;
        if !self.target_node_id.is_empty() {
            struct_ser.serialize_field("targetNodeId", &self.target_node_id)?;
        }
        if self.latency_ms != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("latencyMs", ToString::to_string(&self.latency_ms).as_str())?;
        }
        if let Some(v) = self.measured_at.as_ref() {
            struct_ser.serialize_field("measuredAt", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LatencyMeasurement {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "target_node_id",
            "targetNodeId",
            "latency_ms",
            "latencyMs",
            "measured_at",
            "measuredAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TargetNodeId,
            LatencyMs,
            MeasuredAt,
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
                            "targetNodeId" | "target_node_id" => Ok(GeneratedField::TargetNodeId),
                            "latencyMs" | "latency_ms" => Ok(GeneratedField::LatencyMs),
                            "measuredAt" | "measured_at" => Ok(GeneratedField::MeasuredAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LatencyMeasurement;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.LatencyMeasurement")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LatencyMeasurement, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut target_node_id__ = None;
                let mut latency_ms__ = None;
                let mut measured_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TargetNodeId => {
                            if target_node_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("targetNodeId"));
                            }
                            target_node_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LatencyMs => {
                            if latency_ms__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latencyMs"));
                            }
                            latency_ms__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MeasuredAt => {
                            if measured_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("measuredAt"));
                            }
                            measured_at__ = map_.next_value()?;
                        }
                    }
                }
                Ok(LatencyMeasurement {
                    target_node_id: target_node_id__.unwrap_or_default(),
                    latency_ms: latency_ms__.unwrap_or_default(),
                    measured_at: measured_at__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.LatencyMeasurement", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCancelJob {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.requester_address.is_empty() {
            len += 1;
        }
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgCancelJob", len)?;
        if !self.requester_address.is_empty() {
            struct_ser.serialize_field("requesterAddress", &self.requester_address)?;
        }
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCancelJob {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "requester_address",
            "requesterAddress",
            "job_id",
            "jobId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RequesterAddress,
            JobId,
            Reason,
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
                            "requesterAddress" | "requester_address" => Ok(GeneratedField::RequesterAddress),
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "reason" => Ok(GeneratedField::Reason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCancelJob;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgCancelJob")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCancelJob, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut requester_address__ = None;
                let mut job_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RequesterAddress => {
                            if requester_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requesterAddress"));
                            }
                            requester_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgCancelJob {
                    requester_address: requester_address__.unwrap_or_default(),
                    job_id: job_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgCancelJob", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCancelJobResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgCancelJobResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCancelJobResponse {
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
            type Value = MsgCancelJobResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgCancelJobResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCancelJobResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgCancelJobResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgCancelJobResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateOffering {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.queue_options.is_empty() {
            len += 1;
        }
        if self.pricing.is_some() {
            len += 1;
        }
        if self.required_identity_threshold != 0 {
            len += 1;
        }
        if self.max_runtime_seconds != 0 {
            len += 1;
        }
        if !self.preconfigured_workloads.is_empty() {
            len += 1;
        }
        if self.supports_custom_workloads {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgCreateOffering", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.queue_options.is_empty() {
            struct_ser.serialize_field("queueOptions", &self.queue_options)?;
        }
        if let Some(v) = self.pricing.as_ref() {
            struct_ser.serialize_field("pricing", v)?;
        }
        if self.required_identity_threshold != 0 {
            struct_ser.serialize_field("requiredIdentityThreshold", &self.required_identity_threshold)?;
        }
        if self.max_runtime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntimeSeconds", ToString::to_string(&self.max_runtime_seconds).as_str())?;
        }
        if !self.preconfigured_workloads.is_empty() {
            struct_ser.serialize_field("preconfiguredWorkloads", &self.preconfigured_workloads)?;
        }
        if self.supports_custom_workloads {
            struct_ser.serialize_field("supportsCustomWorkloads", &self.supports_custom_workloads)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateOffering {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "cluster_id",
            "clusterId",
            "name",
            "description",
            "queue_options",
            "queueOptions",
            "pricing",
            "required_identity_threshold",
            "requiredIdentityThreshold",
            "max_runtime_seconds",
            "maxRuntimeSeconds",
            "preconfigured_workloads",
            "preconfiguredWorkloads",
            "supports_custom_workloads",
            "supportsCustomWorkloads",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            ClusterId,
            Name,
            Description,
            QueueOptions,
            Pricing,
            RequiredIdentityThreshold,
            MaxRuntimeSeconds,
            PreconfiguredWorkloads,
            SupportsCustomWorkloads,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "queueOptions" | "queue_options" => Ok(GeneratedField::QueueOptions),
                            "pricing" => Ok(GeneratedField::Pricing),
                            "requiredIdentityThreshold" | "required_identity_threshold" => Ok(GeneratedField::RequiredIdentityThreshold),
                            "maxRuntimeSeconds" | "max_runtime_seconds" => Ok(GeneratedField::MaxRuntimeSeconds),
                            "preconfiguredWorkloads" | "preconfigured_workloads" => Ok(GeneratedField::PreconfiguredWorkloads),
                            "supportsCustomWorkloads" | "supports_custom_workloads" => Ok(GeneratedField::SupportsCustomWorkloads),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateOffering;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgCreateOffering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateOffering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut cluster_id__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut queue_options__ = None;
                let mut pricing__ = None;
                let mut required_identity_threshold__ = None;
                let mut max_runtime_seconds__ = None;
                let mut preconfigured_workloads__ = None;
                let mut supports_custom_workloads__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::QueueOptions => {
                            if queue_options__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queueOptions"));
                            }
                            queue_options__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pricing => {
                            if pricing__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pricing"));
                            }
                            pricing__ = map_.next_value()?;
                        }
                        GeneratedField::RequiredIdentityThreshold => {
                            if required_identity_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredIdentityThreshold"));
                            }
                            required_identity_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRuntimeSeconds => {
                            if max_runtime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntimeSeconds"));
                            }
                            max_runtime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreconfiguredWorkloads => {
                            if preconfigured_workloads__.is_some() {
                                return Err(serde::de::Error::duplicate_field("preconfiguredWorkloads"));
                            }
                            preconfigured_workloads__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SupportsCustomWorkloads => {
                            if supports_custom_workloads__.is_some() {
                                return Err(serde::de::Error::duplicate_field("supportsCustomWorkloads"));
                            }
                            supports_custom_workloads__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgCreateOffering {
                    provider_address: provider_address__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    queue_options: queue_options__.unwrap_or_default(),
                    pricing: pricing__,
                    required_identity_threshold: required_identity_threshold__.unwrap_or_default(),
                    max_runtime_seconds: max_runtime_seconds__.unwrap_or_default(),
                    preconfigured_workloads: preconfigured_workloads__.unwrap_or_default(),
                    supports_custom_workloads: supports_custom_workloads__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgCreateOffering", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateOfferingResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.offering_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgCreateOfferingResponse", len)?;
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateOfferingResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offering_id",
            "offeringId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            OfferingId,
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
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateOfferingResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgCreateOfferingResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateOfferingResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offering_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgCreateOfferingResponse {
                    offering_id: offering_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgCreateOfferingResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeregisterCluster {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgDeregisterCluster", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeregisterCluster {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "cluster_id",
            "clusterId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            ClusterId,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgDeregisterCluster;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgDeregisterCluster")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeregisterCluster, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut cluster_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDeregisterCluster {
                    provider_address: provider_address__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgDeregisterCluster", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeregisterClusterResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgDeregisterClusterResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeregisterClusterResponse {
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
            type Value = MsgDeregisterClusterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgDeregisterClusterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeregisterClusterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDeregisterClusterResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgDeregisterClusterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgFlagDispute {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.disputer_address.is_empty() {
            len += 1;
        }
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.reward_id.is_empty() {
            len += 1;
        }
        if !self.dispute_type.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgFlagDispute", len)?;
        if !self.disputer_address.is_empty() {
            struct_ser.serialize_field("disputerAddress", &self.disputer_address)?;
        }
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.reward_id.is_empty() {
            struct_ser.serialize_field("rewardId", &self.reward_id)?;
        }
        if !self.dispute_type.is_empty() {
            struct_ser.serialize_field("disputeType", &self.dispute_type)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgFlagDispute {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "disputer_address",
            "disputerAddress",
            "job_id",
            "jobId",
            "reward_id",
            "rewardId",
            "dispute_type",
            "disputeType",
            "reason",
            "evidence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DisputerAddress,
            JobId,
            RewardId,
            DisputeType,
            Reason,
            Evidence,
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
                            "disputerAddress" | "disputer_address" => Ok(GeneratedField::DisputerAddress),
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "rewardId" | "reward_id" => Ok(GeneratedField::RewardId),
                            "disputeType" | "dispute_type" => Ok(GeneratedField::DisputeType),
                            "reason" => Ok(GeneratedField::Reason),
                            "evidence" => Ok(GeneratedField::Evidence),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgFlagDispute;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgFlagDispute")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgFlagDispute, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut disputer_address__ = None;
                let mut job_id__ = None;
                let mut reward_id__ = None;
                let mut dispute_type__ = None;
                let mut reason__ = None;
                let mut evidence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DisputerAddress => {
                            if disputer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputerAddress"));
                            }
                            disputer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RewardId => {
                            if reward_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardId"));
                            }
                            reward_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DisputeType => {
                            if dispute_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeType"));
                            }
                            dispute_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgFlagDispute {
                    disputer_address: disputer_address__.unwrap_or_default(),
                    job_id: job_id__.unwrap_or_default(),
                    reward_id: reward_id__.unwrap_or_default(),
                    dispute_type: dispute_type__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgFlagDispute", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgFlagDisputeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.dispute_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgFlagDisputeResponse", len)?;
        if !self.dispute_id.is_empty() {
            struct_ser.serialize_field("disputeId", &self.dispute_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgFlagDisputeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "dispute_id",
            "disputeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DisputeId,
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
                            "disputeId" | "dispute_id" => Ok(GeneratedField::DisputeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgFlagDisputeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgFlagDisputeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgFlagDisputeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut dispute_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DisputeId => {
                            if dispute_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeId"));
                            }
                            dispute_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgFlagDisputeResponse {
                    dispute_id: dispute_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgFlagDisputeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterCluster {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.region.is_empty() {
            len += 1;
        }
        if !self.partitions.is_empty() {
            len += 1;
        }
        if self.total_nodes != 0 {
            len += 1;
        }
        if self.cluster_metadata.is_some() {
            len += 1;
        }
        if !self.slurm_version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgRegisterCluster", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.region.is_empty() {
            struct_ser.serialize_field("region", &self.region)?;
        }
        if !self.partitions.is_empty() {
            struct_ser.serialize_field("partitions", &self.partitions)?;
        }
        if self.total_nodes != 0 {
            struct_ser.serialize_field("totalNodes", &self.total_nodes)?;
        }
        if let Some(v) = self.cluster_metadata.as_ref() {
            struct_ser.serialize_field("clusterMetadata", v)?;
        }
        if !self.slurm_version.is_empty() {
            struct_ser.serialize_field("slurmVersion", &self.slurm_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterCluster {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "name",
            "description",
            "region",
            "partitions",
            "total_nodes",
            "totalNodes",
            "cluster_metadata",
            "clusterMetadata",
            "slurm_version",
            "slurmVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            Name,
            Description,
            Region,
            Partitions,
            TotalNodes,
            ClusterMetadata,
            SlurmVersion,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "region" => Ok(GeneratedField::Region),
                            "partitions" => Ok(GeneratedField::Partitions),
                            "totalNodes" | "total_nodes" => Ok(GeneratedField::TotalNodes),
                            "clusterMetadata" | "cluster_metadata" => Ok(GeneratedField::ClusterMetadata),
                            "slurmVersion" | "slurm_version" => Ok(GeneratedField::SlurmVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterCluster;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgRegisterCluster")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterCluster, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut region__ = None;
                let mut partitions__ = None;
                let mut total_nodes__ = None;
                let mut cluster_metadata__ = None;
                let mut slurm_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Region => {
                            if region__.is_some() {
                                return Err(serde::de::Error::duplicate_field("region"));
                            }
                            region__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Partitions => {
                            if partitions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("partitions"));
                            }
                            partitions__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalNodes => {
                            if total_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalNodes"));
                            }
                            total_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ClusterMetadata => {
                            if cluster_metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterMetadata"));
                            }
                            cluster_metadata__ = map_.next_value()?;
                        }
                        GeneratedField::SlurmVersion => {
                            if slurm_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slurmVersion"));
                            }
                            slurm_version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRegisterCluster {
                    provider_address: provider_address__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    region: region__.unwrap_or_default(),
                    partitions: partitions__.unwrap_or_default(),
                    total_nodes: total_nodes__.unwrap_or_default(),
                    cluster_metadata: cluster_metadata__,
                    slurm_version: slurm_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgRegisterCluster", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterClusterResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgRegisterClusterResponse", len)?;
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterClusterResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster_id",
            "clusterId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClusterId,
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
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterClusterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgRegisterClusterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterClusterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRegisterClusterResponse {
                    cluster_id: cluster_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgRegisterClusterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgReportJobStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.slurm_job_id.is_empty() {
            len += 1;
        }
        if self.state != 0 {
            len += 1;
        }
        if !self.status_message.is_empty() {
            len += 1;
        }
        if self.exit_code != 0 {
            len += 1;
        }
        if self.usage_metrics.is_some() {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        if self.signed_timestamp != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgReportJobStatus", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.slurm_job_id.is_empty() {
            struct_ser.serialize_field("slurmJobId", &self.slurm_job_id)?;
        }
        if self.state != 0 {
            let v = JobState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.status_message.is_empty() {
            struct_ser.serialize_field("statusMessage", &self.status_message)?;
        }
        if self.exit_code != 0 {
            struct_ser.serialize_field("exitCode", &self.exit_code)?;
        }
        if let Some(v) = self.usage_metrics.as_ref() {
            struct_ser.serialize_field("usageMetrics", v)?;
        }
        if !self.signature.is_empty() {
            struct_ser.serialize_field("signature", &self.signature)?;
        }
        if self.signed_timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signedTimestamp", ToString::to_string(&self.signed_timestamp).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgReportJobStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "job_id",
            "jobId",
            "slurm_job_id",
            "slurmJobId",
            "state",
            "status_message",
            "statusMessage",
            "exit_code",
            "exitCode",
            "usage_metrics",
            "usageMetrics",
            "signature",
            "signed_timestamp",
            "signedTimestamp",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            JobId,
            SlurmJobId,
            State,
            StatusMessage,
            ExitCode,
            UsageMetrics,
            Signature,
            SignedTimestamp,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "slurmJobId" | "slurm_job_id" => Ok(GeneratedField::SlurmJobId),
                            "state" => Ok(GeneratedField::State),
                            "statusMessage" | "status_message" => Ok(GeneratedField::StatusMessage),
                            "exitCode" | "exit_code" => Ok(GeneratedField::ExitCode),
                            "usageMetrics" | "usage_metrics" => Ok(GeneratedField::UsageMetrics),
                            "signature" => Ok(GeneratedField::Signature),
                            "signedTimestamp" | "signed_timestamp" => Ok(GeneratedField::SignedTimestamp),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgReportJobStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgReportJobStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgReportJobStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut job_id__ = None;
                let mut slurm_job_id__ = None;
                let mut state__ = None;
                let mut status_message__ = None;
                let mut exit_code__ = None;
                let mut usage_metrics__ = None;
                let mut signature__ = None;
                let mut signed_timestamp__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SlurmJobId => {
                            if slurm_job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slurmJobId"));
                            }
                            slurm_job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<JobState>()? as i32);
                        }
                        GeneratedField::StatusMessage => {
                            if status_message__.is_some() {
                                return Err(serde::de::Error::duplicate_field("statusMessage"));
                            }
                            status_message__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExitCode => {
                            if exit_code__.is_some() {
                                return Err(serde::de::Error::duplicate_field("exitCode"));
                            }
                            exit_code__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UsageMetrics => {
                            if usage_metrics__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageMetrics"));
                            }
                            usage_metrics__ = map_.next_value()?;
                        }
                        GeneratedField::Signature => {
                            if signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signature"));
                            }
                            signature__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SignedTimestamp => {
                            if signed_timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signedTimestamp"));
                            }
                            signed_timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgReportJobStatus {
                    provider_address: provider_address__.unwrap_or_default(),
                    job_id: job_id__.unwrap_or_default(),
                    slurm_job_id: slurm_job_id__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    status_message: status_message__.unwrap_or_default(),
                    exit_code: exit_code__.unwrap_or_default(),
                    usage_metrics: usage_metrics__,
                    signature: signature__.unwrap_or_default(),
                    signed_timestamp: signed_timestamp__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgReportJobStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgReportJobStatusResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgReportJobStatusResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgReportJobStatusResponse {
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
            type Value = MsgReportJobStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgReportJobStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgReportJobStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgReportJobStatusResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgReportJobStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveDispute {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.resolver_address.is_empty() {
            len += 1;
        }
        if !self.dispute_id.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.resolution.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgResolveDispute", len)?;
        if !self.resolver_address.is_empty() {
            struct_ser.serialize_field("resolverAddress", &self.resolver_address)?;
        }
        if !self.dispute_id.is_empty() {
            struct_ser.serialize_field("disputeId", &self.dispute_id)?;
        }
        if self.status != 0 {
            let v = DisputeStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.resolution.is_empty() {
            struct_ser.serialize_field("resolution", &self.resolution)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveDispute {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "resolver_address",
            "resolverAddress",
            "dispute_id",
            "disputeId",
            "status",
            "resolution",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ResolverAddress,
            DisputeId,
            Status,
            Resolution,
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
                            "resolverAddress" | "resolver_address" => Ok(GeneratedField::ResolverAddress),
                            "disputeId" | "dispute_id" => Ok(GeneratedField::DisputeId),
                            "status" => Ok(GeneratedField::Status),
                            "resolution" => Ok(GeneratedField::Resolution),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgResolveDispute;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgResolveDispute")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveDispute, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut resolver_address__ = None;
                let mut dispute_id__ = None;
                let mut status__ = None;
                let mut resolution__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ResolverAddress => {
                            if resolver_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolverAddress"));
                            }
                            resolver_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DisputeId => {
                            if dispute_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeId"));
                            }
                            dispute_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<DisputeStatus>()? as i32);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgResolveDispute {
                    resolver_address: resolver_address__.unwrap_or_default(),
                    dispute_id: dispute_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgResolveDispute", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveDisputeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgResolveDisputeResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveDisputeResponse {
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
            type Value = MsgResolveDisputeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgResolveDisputeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveDisputeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgResolveDisputeResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgResolveDisputeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitJob {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.customer_address.is_empty() {
            len += 1;
        }
        if !self.offering_id.is_empty() {
            len += 1;
        }
        if !self.queue_name.is_empty() {
            len += 1;
        }
        if self.workload_spec.is_some() {
            len += 1;
        }
        if self.resources.is_some() {
            len += 1;
        }
        if !self.data_references.is_empty() {
            len += 1;
        }
        if !self.encrypted_inputs_pointer.is_empty() {
            len += 1;
        }
        if !self.encrypted_outputs_pointer.is_empty() {
            len += 1;
        }
        if self.max_runtime_seconds != 0 {
            len += 1;
        }
        if !self.max_price.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgSubmitJob", len)?;
        if !self.customer_address.is_empty() {
            struct_ser.serialize_field("customerAddress", &self.customer_address)?;
        }
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if !self.queue_name.is_empty() {
            struct_ser.serialize_field("queueName", &self.queue_name)?;
        }
        if let Some(v) = self.workload_spec.as_ref() {
            struct_ser.serialize_field("workloadSpec", v)?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        if !self.data_references.is_empty() {
            struct_ser.serialize_field("dataReferences", &self.data_references)?;
        }
        if !self.encrypted_inputs_pointer.is_empty() {
            struct_ser.serialize_field("encryptedInputsPointer", &self.encrypted_inputs_pointer)?;
        }
        if !self.encrypted_outputs_pointer.is_empty() {
            struct_ser.serialize_field("encryptedOutputsPointer", &self.encrypted_outputs_pointer)?;
        }
        if self.max_runtime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntimeSeconds", ToString::to_string(&self.max_runtime_seconds).as_str())?;
        }
        if !self.max_price.is_empty() {
            struct_ser.serialize_field("maxPrice", &self.max_price)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitJob {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "customer_address",
            "customerAddress",
            "offering_id",
            "offeringId",
            "queue_name",
            "queueName",
            "workload_spec",
            "workloadSpec",
            "resources",
            "data_references",
            "dataReferences",
            "encrypted_inputs_pointer",
            "encryptedInputsPointer",
            "encrypted_outputs_pointer",
            "encryptedOutputsPointer",
            "max_runtime_seconds",
            "maxRuntimeSeconds",
            "max_price",
            "maxPrice",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CustomerAddress,
            OfferingId,
            QueueName,
            WorkloadSpec,
            Resources,
            DataReferences,
            EncryptedInputsPointer,
            EncryptedOutputsPointer,
            MaxRuntimeSeconds,
            MaxPrice,
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
                            "customerAddress" | "customer_address" => Ok(GeneratedField::CustomerAddress),
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            "queueName" | "queue_name" => Ok(GeneratedField::QueueName),
                            "workloadSpec" | "workload_spec" => Ok(GeneratedField::WorkloadSpec),
                            "resources" => Ok(GeneratedField::Resources),
                            "dataReferences" | "data_references" => Ok(GeneratedField::DataReferences),
                            "encryptedInputsPointer" | "encrypted_inputs_pointer" => Ok(GeneratedField::EncryptedInputsPointer),
                            "encryptedOutputsPointer" | "encrypted_outputs_pointer" => Ok(GeneratedField::EncryptedOutputsPointer),
                            "maxRuntimeSeconds" | "max_runtime_seconds" => Ok(GeneratedField::MaxRuntimeSeconds),
                            "maxPrice" | "max_price" => Ok(GeneratedField::MaxPrice),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitJob;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgSubmitJob")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitJob, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut customer_address__ = None;
                let mut offering_id__ = None;
                let mut queue_name__ = None;
                let mut workload_spec__ = None;
                let mut resources__ = None;
                let mut data_references__ = None;
                let mut encrypted_inputs_pointer__ = None;
                let mut encrypted_outputs_pointer__ = None;
                let mut max_runtime_seconds__ = None;
                let mut max_price__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CustomerAddress => {
                            if customer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customerAddress"));
                            }
                            customer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::QueueName => {
                            if queue_name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queueName"));
                            }
                            queue_name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::WorkloadSpec => {
                            if workload_spec__.is_some() {
                                return Err(serde::de::Error::duplicate_field("workloadSpec"));
                            }
                            workload_spec__ = map_.next_value()?;
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                        GeneratedField::DataReferences => {
                            if data_references__.is_some() {
                                return Err(serde::de::Error::duplicate_field("dataReferences"));
                            }
                            data_references__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptedInputsPointer => {
                            if encrypted_inputs_pointer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedInputsPointer"));
                            }
                            encrypted_inputs_pointer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EncryptedOutputsPointer => {
                            if encrypted_outputs_pointer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedOutputsPointer"));
                            }
                            encrypted_outputs_pointer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxRuntimeSeconds => {
                            if max_runtime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntimeSeconds"));
                            }
                            max_runtime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxPrice => {
                            if max_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxPrice"));
                            }
                            max_price__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitJob {
                    customer_address: customer_address__.unwrap_or_default(),
                    offering_id: offering_id__.unwrap_or_default(),
                    queue_name: queue_name__.unwrap_or_default(),
                    workload_spec: workload_spec__,
                    resources: resources__,
                    data_references: data_references__.unwrap_or_default(),
                    encrypted_inputs_pointer: encrypted_inputs_pointer__.unwrap_or_default(),
                    encrypted_outputs_pointer: encrypted_outputs_pointer__.unwrap_or_default(),
                    max_runtime_seconds: max_runtime_seconds__.unwrap_or_default(),
                    max_price: max_price__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgSubmitJob", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitJobResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.escrow_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgSubmitJobResponse", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.escrow_id.is_empty() {
            struct_ser.serialize_field("escrowId", &self.escrow_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitJobResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
            "escrow_id",
            "escrowId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
            EscrowId,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "escrowId" | "escrow_id" => Ok(GeneratedField::EscrowId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitJobResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgSubmitJobResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitJobResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                let mut escrow_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowId => {
                            if escrow_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowId"));
                            }
                            escrow_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitJobResponse {
                    job_id: job_id__.unwrap_or_default(),
                    escrow_id: escrow_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgSubmitJobResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateCluster {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.state != 0 {
            len += 1;
        }
        if !self.partitions.is_empty() {
            len += 1;
        }
        if self.total_nodes != 0 {
            len += 1;
        }
        if self.available_nodes != 0 {
            len += 1;
        }
        if self.cluster_metadata.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateCluster", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.state != 0 {
            let v = ClusterState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.partitions.is_empty() {
            struct_ser.serialize_field("partitions", &self.partitions)?;
        }
        if self.total_nodes != 0 {
            struct_ser.serialize_field("totalNodes", &self.total_nodes)?;
        }
        if self.available_nodes != 0 {
            struct_ser.serialize_field("availableNodes", &self.available_nodes)?;
        }
        if let Some(v) = self.cluster_metadata.as_ref() {
            struct_ser.serialize_field("clusterMetadata", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateCluster {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "cluster_id",
            "clusterId",
            "name",
            "description",
            "state",
            "partitions",
            "total_nodes",
            "totalNodes",
            "available_nodes",
            "availableNodes",
            "cluster_metadata",
            "clusterMetadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            ClusterId,
            Name,
            Description,
            State,
            Partitions,
            TotalNodes,
            AvailableNodes,
            ClusterMetadata,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "state" => Ok(GeneratedField::State),
                            "partitions" => Ok(GeneratedField::Partitions),
                            "totalNodes" | "total_nodes" => Ok(GeneratedField::TotalNodes),
                            "availableNodes" | "available_nodes" => Ok(GeneratedField::AvailableNodes),
                            "clusterMetadata" | "cluster_metadata" => Ok(GeneratedField::ClusterMetadata),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateCluster;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateCluster")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateCluster, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut cluster_id__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut state__ = None;
                let mut partitions__ = None;
                let mut total_nodes__ = None;
                let mut available_nodes__ = None;
                let mut cluster_metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<ClusterState>()? as i32);
                        }
                        GeneratedField::Partitions => {
                            if partitions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("partitions"));
                            }
                            partitions__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalNodes => {
                            if total_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalNodes"));
                            }
                            total_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AvailableNodes => {
                            if available_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("availableNodes"));
                            }
                            available_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ClusterMetadata => {
                            if cluster_metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterMetadata"));
                            }
                            cluster_metadata__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgUpdateCluster {
                    provider_address: provider_address__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    partitions: partitions__.unwrap_or_default(),
                    total_nodes: total_nodes__.unwrap_or_default(),
                    available_nodes: available_nodes__.unwrap_or_default(),
                    cluster_metadata: cluster_metadata__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateCluster", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateClusterResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateClusterResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateClusterResponse {
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
            type Value = MsgUpdateClusterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateClusterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateClusterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateClusterResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateClusterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateNodeMetadata {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.node_id.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.region.is_empty() {
            len += 1;
        }
        if !self.datacenter.is_empty() {
            len += 1;
        }
        if !self.latency_measurements.is_empty() {
            len += 1;
        }
        if self.network_bandwidth_mbps != 0 {
            len += 1;
        }
        if self.resources.is_some() {
            len += 1;
        }
        if self.active {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateNodeMetadata", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.node_id.is_empty() {
            struct_ser.serialize_field("nodeId", &self.node_id)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.region.is_empty() {
            struct_ser.serialize_field("region", &self.region)?;
        }
        if !self.datacenter.is_empty() {
            struct_ser.serialize_field("datacenter", &self.datacenter)?;
        }
        if !self.latency_measurements.is_empty() {
            struct_ser.serialize_field("latencyMeasurements", &self.latency_measurements)?;
        }
        if self.network_bandwidth_mbps != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("networkBandwidthMbps", ToString::to_string(&self.network_bandwidth_mbps).as_str())?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        if self.active {
            struct_ser.serialize_field("active", &self.active)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateNodeMetadata {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "node_id",
            "nodeId",
            "cluster_id",
            "clusterId",
            "region",
            "datacenter",
            "latency_measurements",
            "latencyMeasurements",
            "network_bandwidth_mbps",
            "networkBandwidthMbps",
            "resources",
            "active",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            NodeId,
            ClusterId,
            Region,
            Datacenter,
            LatencyMeasurements,
            NetworkBandwidthMbps,
            Resources,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "nodeId" | "node_id" => Ok(GeneratedField::NodeId),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "region" => Ok(GeneratedField::Region),
                            "datacenter" => Ok(GeneratedField::Datacenter),
                            "latencyMeasurements" | "latency_measurements" => Ok(GeneratedField::LatencyMeasurements),
                            "networkBandwidthMbps" | "network_bandwidth_mbps" => Ok(GeneratedField::NetworkBandwidthMbps),
                            "resources" => Ok(GeneratedField::Resources),
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
            type Value = MsgUpdateNodeMetadata;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateNodeMetadata")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateNodeMetadata, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut node_id__ = None;
                let mut cluster_id__ = None;
                let mut region__ = None;
                let mut datacenter__ = None;
                let mut latency_measurements__ = None;
                let mut network_bandwidth_mbps__ = None;
                let mut resources__ = None;
                let mut active__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NodeId => {
                            if node_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeId"));
                            }
                            node_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Region => {
                            if region__.is_some() {
                                return Err(serde::de::Error::duplicate_field("region"));
                            }
                            region__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Datacenter => {
                            if datacenter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("datacenter"));
                            }
                            datacenter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LatencyMeasurements => {
                            if latency_measurements__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latencyMeasurements"));
                            }
                            latency_measurements__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NetworkBandwidthMbps => {
                            if network_bandwidth_mbps__.is_some() {
                                return Err(serde::de::Error::duplicate_field("networkBandwidthMbps"));
                            }
                            network_bandwidth_mbps__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateNodeMetadata {
                    provider_address: provider_address__.unwrap_or_default(),
                    node_id: node_id__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    region: region__.unwrap_or_default(),
                    datacenter: datacenter__.unwrap_or_default(),
                    latency_measurements: latency_measurements__.unwrap_or_default(),
                    network_bandwidth_mbps: network_bandwidth_mbps__.unwrap_or_default(),
                    resources: resources__,
                    active: active__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateNodeMetadata", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateNodeMetadataResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateNodeMetadataResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateNodeMetadataResponse {
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
            type Value = MsgUpdateNodeMetadataResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateNodeMetadataResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateNodeMetadataResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateNodeMetadataResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateNodeMetadataResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateOffering {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.offering_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.queue_options.is_empty() {
            len += 1;
        }
        if self.pricing.is_some() {
            len += 1;
        }
        if self.required_identity_threshold != 0 {
            len += 1;
        }
        if self.max_runtime_seconds != 0 {
            len += 1;
        }
        if !self.preconfigured_workloads.is_empty() {
            len += 1;
        }
        if self.supports_custom_workloads {
            len += 1;
        }
        if self.active {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateOffering", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.queue_options.is_empty() {
            struct_ser.serialize_field("queueOptions", &self.queue_options)?;
        }
        if let Some(v) = self.pricing.as_ref() {
            struct_ser.serialize_field("pricing", v)?;
        }
        if self.required_identity_threshold != 0 {
            struct_ser.serialize_field("requiredIdentityThreshold", &self.required_identity_threshold)?;
        }
        if self.max_runtime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntimeSeconds", ToString::to_string(&self.max_runtime_seconds).as_str())?;
        }
        if !self.preconfigured_workloads.is_empty() {
            struct_ser.serialize_field("preconfiguredWorkloads", &self.preconfigured_workloads)?;
        }
        if self.supports_custom_workloads {
            struct_ser.serialize_field("supportsCustomWorkloads", &self.supports_custom_workloads)?;
        }
        if self.active {
            struct_ser.serialize_field("active", &self.active)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateOffering {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "offering_id",
            "offeringId",
            "name",
            "description",
            "queue_options",
            "queueOptions",
            "pricing",
            "required_identity_threshold",
            "requiredIdentityThreshold",
            "max_runtime_seconds",
            "maxRuntimeSeconds",
            "preconfigured_workloads",
            "preconfiguredWorkloads",
            "supports_custom_workloads",
            "supportsCustomWorkloads",
            "active",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            OfferingId,
            Name,
            Description,
            QueueOptions,
            Pricing,
            RequiredIdentityThreshold,
            MaxRuntimeSeconds,
            PreconfiguredWorkloads,
            SupportsCustomWorkloads,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "queueOptions" | "queue_options" => Ok(GeneratedField::QueueOptions),
                            "pricing" => Ok(GeneratedField::Pricing),
                            "requiredIdentityThreshold" | "required_identity_threshold" => Ok(GeneratedField::RequiredIdentityThreshold),
                            "maxRuntimeSeconds" | "max_runtime_seconds" => Ok(GeneratedField::MaxRuntimeSeconds),
                            "preconfiguredWorkloads" | "preconfigured_workloads" => Ok(GeneratedField::PreconfiguredWorkloads),
                            "supportsCustomWorkloads" | "supports_custom_workloads" => Ok(GeneratedField::SupportsCustomWorkloads),
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
            type Value = MsgUpdateOffering;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateOffering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateOffering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut offering_id__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut queue_options__ = None;
                let mut pricing__ = None;
                let mut required_identity_threshold__ = None;
                let mut max_runtime_seconds__ = None;
                let mut preconfigured_workloads__ = None;
                let mut supports_custom_workloads__ = None;
                let mut active__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::QueueOptions => {
                            if queue_options__.is_some() {
                                return Err(serde::de::Error::duplicate_field("queueOptions"));
                            }
                            queue_options__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pricing => {
                            if pricing__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pricing"));
                            }
                            pricing__ = map_.next_value()?;
                        }
                        GeneratedField::RequiredIdentityThreshold => {
                            if required_identity_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredIdentityThreshold"));
                            }
                            required_identity_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRuntimeSeconds => {
                            if max_runtime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntimeSeconds"));
                            }
                            max_runtime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreconfiguredWorkloads => {
                            if preconfigured_workloads__.is_some() {
                                return Err(serde::de::Error::duplicate_field("preconfiguredWorkloads"));
                            }
                            preconfigured_workloads__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SupportsCustomWorkloads => {
                            if supports_custom_workloads__.is_some() {
                                return Err(serde::de::Error::duplicate_field("supportsCustomWorkloads"));
                            }
                            supports_custom_workloads__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateOffering {
                    provider_address: provider_address__.unwrap_or_default(),
                    offering_id: offering_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    queue_options: queue_options__.unwrap_or_default(),
                    pricing: pricing__,
                    required_identity_threshold: required_identity_threshold__.unwrap_or_default(),
                    max_runtime_seconds: max_runtime_seconds__.unwrap_or_default(),
                    preconfigured_workloads: preconfigured_workloads__.unwrap_or_default(),
                    supports_custom_workloads: supports_custom_workloads__.unwrap_or_default(),
                    active: active__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateOffering", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateOfferingResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateOfferingResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateOfferingResponse {
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
            type Value = MsgUpdateOfferingResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateOfferingResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateOfferingResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateOfferingResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateOfferingResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.hpc.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.hpc.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for NodeMetadata {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.node_id.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.region.is_empty() {
            len += 1;
        }
        if !self.datacenter.is_empty() {
            len += 1;
        }
        if !self.latency_measurements.is_empty() {
            len += 1;
        }
        if self.avg_latency_ms != 0 {
            len += 1;
        }
        if self.network_bandwidth_mbps != 0 {
            len += 1;
        }
        if self.resources.is_some() {
            len += 1;
        }
        if self.active {
            len += 1;
        }
        if self.last_heartbeat.is_some() {
            len += 1;
        }
        if self.joined_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.NodeMetadata", len)?;
        if !self.node_id.is_empty() {
            struct_ser.serialize_field("nodeId", &self.node_id)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.region.is_empty() {
            struct_ser.serialize_field("region", &self.region)?;
        }
        if !self.datacenter.is_empty() {
            struct_ser.serialize_field("datacenter", &self.datacenter)?;
        }
        if !self.latency_measurements.is_empty() {
            struct_ser.serialize_field("latencyMeasurements", &self.latency_measurements)?;
        }
        if self.avg_latency_ms != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("avgLatencyMs", ToString::to_string(&self.avg_latency_ms).as_str())?;
        }
        if self.network_bandwidth_mbps != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("networkBandwidthMbps", ToString::to_string(&self.network_bandwidth_mbps).as_str())?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        if self.active {
            struct_ser.serialize_field("active", &self.active)?;
        }
        if let Some(v) = self.last_heartbeat.as_ref() {
            struct_ser.serialize_field("lastHeartbeat", v)?;
        }
        if let Some(v) = self.joined_at.as_ref() {
            struct_ser.serialize_field("joinedAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for NodeMetadata {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "node_id",
            "nodeId",
            "cluster_id",
            "clusterId",
            "provider_address",
            "providerAddress",
            "region",
            "datacenter",
            "latency_measurements",
            "latencyMeasurements",
            "avg_latency_ms",
            "avgLatencyMs",
            "network_bandwidth_mbps",
            "networkBandwidthMbps",
            "resources",
            "active",
            "last_heartbeat",
            "lastHeartbeat",
            "joined_at",
            "joinedAt",
            "updated_at",
            "updatedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            NodeId,
            ClusterId,
            ProviderAddress,
            Region,
            Datacenter,
            LatencyMeasurements,
            AvgLatencyMs,
            NetworkBandwidthMbps,
            Resources,
            Active,
            LastHeartbeat,
            JoinedAt,
            UpdatedAt,
            BlockHeight,
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
                            "nodeId" | "node_id" => Ok(GeneratedField::NodeId),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "region" => Ok(GeneratedField::Region),
                            "datacenter" => Ok(GeneratedField::Datacenter),
                            "latencyMeasurements" | "latency_measurements" => Ok(GeneratedField::LatencyMeasurements),
                            "avgLatencyMs" | "avg_latency_ms" => Ok(GeneratedField::AvgLatencyMs),
                            "networkBandwidthMbps" | "network_bandwidth_mbps" => Ok(GeneratedField::NetworkBandwidthMbps),
                            "resources" => Ok(GeneratedField::Resources),
                            "active" => Ok(GeneratedField::Active),
                            "lastHeartbeat" | "last_heartbeat" => Ok(GeneratedField::LastHeartbeat),
                            "joinedAt" | "joined_at" => Ok(GeneratedField::JoinedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = NodeMetadata;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.NodeMetadata")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<NodeMetadata, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut node_id__ = None;
                let mut cluster_id__ = None;
                let mut provider_address__ = None;
                let mut region__ = None;
                let mut datacenter__ = None;
                let mut latency_measurements__ = None;
                let mut avg_latency_ms__ = None;
                let mut network_bandwidth_mbps__ = None;
                let mut resources__ = None;
                let mut active__ = None;
                let mut last_heartbeat__ = None;
                let mut joined_at__ = None;
                let mut updated_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::NodeId => {
                            if node_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeId"));
                            }
                            node_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Region => {
                            if region__.is_some() {
                                return Err(serde::de::Error::duplicate_field("region"));
                            }
                            region__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Datacenter => {
                            if datacenter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("datacenter"));
                            }
                            datacenter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LatencyMeasurements => {
                            if latency_measurements__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latencyMeasurements"));
                            }
                            latency_measurements__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AvgLatencyMs => {
                            if avg_latency_ms__.is_some() {
                                return Err(serde::de::Error::duplicate_field("avgLatencyMs"));
                            }
                            avg_latency_ms__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NetworkBandwidthMbps => {
                            if network_bandwidth_mbps__.is_some() {
                                return Err(serde::de::Error::duplicate_field("networkBandwidthMbps"));
                            }
                            network_bandwidth_mbps__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastHeartbeat => {
                            if last_heartbeat__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastHeartbeat"));
                            }
                            last_heartbeat__ = map_.next_value()?;
                        }
                        GeneratedField::JoinedAt => {
                            if joined_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("joinedAt"));
                            }
                            joined_at__ = map_.next_value()?;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(NodeMetadata {
                    node_id: node_id__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    region: region__.unwrap_or_default(),
                    datacenter: datacenter__.unwrap_or_default(),
                    latency_measurements: latency_measurements__.unwrap_or_default(),
                    avg_latency_ms: avg_latency_ms__.unwrap_or_default(),
                    network_bandwidth_mbps: network_bandwidth_mbps__.unwrap_or_default(),
                    resources: resources__,
                    active: active__.unwrap_or_default(),
                    last_heartbeat: last_heartbeat__,
                    joined_at: joined_at__,
                    updated_at: updated_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.NodeMetadata", FIELDS, GeneratedVisitor)
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
        if self.cpu_cores != 0 {
            len += 1;
        }
        if self.memory_gb != 0 {
            len += 1;
        }
        if self.gpus != 0 {
            len += 1;
        }
        if !self.gpu_type.is_empty() {
            len += 1;
        }
        if self.storage_gb != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.NodeResources", len)?;
        if self.cpu_cores != 0 {
            struct_ser.serialize_field("cpuCores", &self.cpu_cores)?;
        }
        if self.memory_gb != 0 {
            struct_ser.serialize_field("memoryGb", &self.memory_gb)?;
        }
        if self.gpus != 0 {
            struct_ser.serialize_field("gpus", &self.gpus)?;
        }
        if !self.gpu_type.is_empty() {
            struct_ser.serialize_field("gpuType", &self.gpu_type)?;
        }
        if self.storage_gb != 0 {
            struct_ser.serialize_field("storageGb", &self.storage_gb)?;
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
            "cpu_cores",
            "cpuCores",
            "memory_gb",
            "memoryGb",
            "gpus",
            "gpu_type",
            "gpuType",
            "storage_gb",
            "storageGb",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CpuCores,
            MemoryGb,
            Gpus,
            GpuType,
            StorageGb,
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
                            "cpuCores" | "cpu_cores" => Ok(GeneratedField::CpuCores),
                            "memoryGb" | "memory_gb" => Ok(GeneratedField::MemoryGb),
                            "gpus" => Ok(GeneratedField::Gpus),
                            "gpuType" | "gpu_type" => Ok(GeneratedField::GpuType),
                            "storageGb" | "storage_gb" => Ok(GeneratedField::StorageGb),
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
                formatter.write_str("struct virtengine.hpc.v1.NodeResources")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<NodeResources, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cpu_cores__ = None;
                let mut memory_gb__ = None;
                let mut gpus__ = None;
                let mut gpu_type__ = None;
                let mut storage_gb__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CpuCores => {
                            if cpu_cores__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cpuCores"));
                            }
                            cpu_cores__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MemoryGb => {
                            if memory_gb__.is_some() {
                                return Err(serde::de::Error::duplicate_field("memoryGb"));
                            }
                            memory_gb__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Gpus => {
                            if gpus__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpus"));
                            }
                            gpus__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GpuType => {
                            if gpu_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("gpuType"));
                            }
                            gpu_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StorageGb => {
                            if storage_gb__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storageGb"));
                            }
                            storage_gb__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(NodeResources {
                    cpu_cores: cpu_cores__.unwrap_or_default(),
                    memory_gb: memory_gb__.unwrap_or_default(),
                    gpus: gpus__.unwrap_or_default(),
                    gpu_type: gpu_type__.unwrap_or_default(),
                    storage_gb: storage_gb__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.NodeResources", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for NodeReward {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.node_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.amount.is_empty() {
            len += 1;
        }
        if !self.contribution_weight.is_empty() {
            len += 1;
        }
        if self.usage_seconds != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.NodeReward", len)?;
        if !self.node_id.is_empty() {
            struct_ser.serialize_field("nodeId", &self.node_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        if !self.contribution_weight.is_empty() {
            struct_ser.serialize_field("contributionWeight", &self.contribution_weight)?;
        }
        if self.usage_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("usageSeconds", ToString::to_string(&self.usage_seconds).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for NodeReward {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "node_id",
            "nodeId",
            "provider_address",
            "providerAddress",
            "amount",
            "contribution_weight",
            "contributionWeight",
            "usage_seconds",
            "usageSeconds",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            NodeId,
            ProviderAddress,
            Amount,
            ContributionWeight,
            UsageSeconds,
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
                            "nodeId" | "node_id" => Ok(GeneratedField::NodeId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "amount" => Ok(GeneratedField::Amount),
                            "contributionWeight" | "contribution_weight" => Ok(GeneratedField::ContributionWeight),
                            "usageSeconds" | "usage_seconds" => Ok(GeneratedField::UsageSeconds),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = NodeReward;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.NodeReward")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<NodeReward, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut node_id__ = None;
                let mut provider_address__ = None;
                let mut amount__ = None;
                let mut contribution_weight__ = None;
                let mut usage_seconds__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::NodeId => {
                            if node_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeId"));
                            }
                            node_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ContributionWeight => {
                            if contribution_weight__.is_some() {
                                return Err(serde::de::Error::duplicate_field("contributionWeight"));
                            }
                            contribution_weight__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UsageSeconds => {
                            if usage_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageSeconds"));
                            }
                            usage_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(NodeReward {
                    node_id: node_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    amount: amount__.unwrap_or_default(),
                    contribution_weight: contribution_weight__.unwrap_or_default(),
                    usage_seconds: usage_seconds__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.NodeReward", FIELDS, GeneratedVisitor)
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
        if !self.platform_fee_rate.is_empty() {
            len += 1;
        }
        if !self.provider_reward_rate.is_empty() {
            len += 1;
        }
        if !self.node_reward_rate.is_empty() {
            len += 1;
        }
        if self.min_job_duration_seconds != 0 {
            len += 1;
        }
        if self.max_job_duration_seconds != 0 {
            len += 1;
        }
        if self.default_identity_threshold != 0 {
            len += 1;
        }
        if self.cluster_heartbeat_timeout != 0 {
            len += 1;
        }
        if self.node_heartbeat_timeout != 0 {
            len += 1;
        }
        if !self.latency_weight_factor.is_empty() {
            len += 1;
        }
        if !self.capacity_weight_factor.is_empty() {
            len += 1;
        }
        if self.max_latency_ms != 0 {
            len += 1;
        }
        if self.dispute_resolution_period != 0 {
            len += 1;
        }
        if !self.reward_formula_version.is_empty() {
            len += 1;
        }
        if self.enable_proximity_clustering {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.Params", len)?;
        if !self.platform_fee_rate.is_empty() {
            struct_ser.serialize_field("platformFeeRate", &self.platform_fee_rate)?;
        }
        if !self.provider_reward_rate.is_empty() {
            struct_ser.serialize_field("providerRewardRate", &self.provider_reward_rate)?;
        }
        if !self.node_reward_rate.is_empty() {
            struct_ser.serialize_field("nodeRewardRate", &self.node_reward_rate)?;
        }
        if self.min_job_duration_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("minJobDurationSeconds", ToString::to_string(&self.min_job_duration_seconds).as_str())?;
        }
        if self.max_job_duration_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxJobDurationSeconds", ToString::to_string(&self.max_job_duration_seconds).as_str())?;
        }
        if self.default_identity_threshold != 0 {
            struct_ser.serialize_field("defaultIdentityThreshold", &self.default_identity_threshold)?;
        }
        if self.cluster_heartbeat_timeout != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("clusterHeartbeatTimeout", ToString::to_string(&self.cluster_heartbeat_timeout).as_str())?;
        }
        if self.node_heartbeat_timeout != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nodeHeartbeatTimeout", ToString::to_string(&self.node_heartbeat_timeout).as_str())?;
        }
        if !self.latency_weight_factor.is_empty() {
            struct_ser.serialize_field("latencyWeightFactor", &self.latency_weight_factor)?;
        }
        if !self.capacity_weight_factor.is_empty() {
            struct_ser.serialize_field("capacityWeightFactor", &self.capacity_weight_factor)?;
        }
        if self.max_latency_ms != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxLatencyMs", ToString::to_string(&self.max_latency_ms).as_str())?;
        }
        if self.dispute_resolution_period != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("disputeResolutionPeriod", ToString::to_string(&self.dispute_resolution_period).as_str())?;
        }
        if !self.reward_formula_version.is_empty() {
            struct_ser.serialize_field("rewardFormulaVersion", &self.reward_formula_version)?;
        }
        if self.enable_proximity_clustering {
            struct_ser.serialize_field("enableProximityClustering", &self.enable_proximity_clustering)?;
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
            "platform_fee_rate",
            "platformFeeRate",
            "provider_reward_rate",
            "providerRewardRate",
            "node_reward_rate",
            "nodeRewardRate",
            "min_job_duration_seconds",
            "minJobDurationSeconds",
            "max_job_duration_seconds",
            "maxJobDurationSeconds",
            "default_identity_threshold",
            "defaultIdentityThreshold",
            "cluster_heartbeat_timeout",
            "clusterHeartbeatTimeout",
            "node_heartbeat_timeout",
            "nodeHeartbeatTimeout",
            "latency_weight_factor",
            "latencyWeightFactor",
            "capacity_weight_factor",
            "capacityWeightFactor",
            "max_latency_ms",
            "maxLatencyMs",
            "dispute_resolution_period",
            "disputeResolutionPeriod",
            "reward_formula_version",
            "rewardFormulaVersion",
            "enable_proximity_clustering",
            "enableProximityClustering",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            PlatformFeeRate,
            ProviderRewardRate,
            NodeRewardRate,
            MinJobDurationSeconds,
            MaxJobDurationSeconds,
            DefaultIdentityThreshold,
            ClusterHeartbeatTimeout,
            NodeHeartbeatTimeout,
            LatencyWeightFactor,
            CapacityWeightFactor,
            MaxLatencyMs,
            DisputeResolutionPeriod,
            RewardFormulaVersion,
            EnableProximityClustering,
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
                            "platformFeeRate" | "platform_fee_rate" => Ok(GeneratedField::PlatformFeeRate),
                            "providerRewardRate" | "provider_reward_rate" => Ok(GeneratedField::ProviderRewardRate),
                            "nodeRewardRate" | "node_reward_rate" => Ok(GeneratedField::NodeRewardRate),
                            "minJobDurationSeconds" | "min_job_duration_seconds" => Ok(GeneratedField::MinJobDurationSeconds),
                            "maxJobDurationSeconds" | "max_job_duration_seconds" => Ok(GeneratedField::MaxJobDurationSeconds),
                            "defaultIdentityThreshold" | "default_identity_threshold" => Ok(GeneratedField::DefaultIdentityThreshold),
                            "clusterHeartbeatTimeout" | "cluster_heartbeat_timeout" => Ok(GeneratedField::ClusterHeartbeatTimeout),
                            "nodeHeartbeatTimeout" | "node_heartbeat_timeout" => Ok(GeneratedField::NodeHeartbeatTimeout),
                            "latencyWeightFactor" | "latency_weight_factor" => Ok(GeneratedField::LatencyWeightFactor),
                            "capacityWeightFactor" | "capacity_weight_factor" => Ok(GeneratedField::CapacityWeightFactor),
                            "maxLatencyMs" | "max_latency_ms" => Ok(GeneratedField::MaxLatencyMs),
                            "disputeResolutionPeriod" | "dispute_resolution_period" => Ok(GeneratedField::DisputeResolutionPeriod),
                            "rewardFormulaVersion" | "reward_formula_version" => Ok(GeneratedField::RewardFormulaVersion),
                            "enableProximityClustering" | "enable_proximity_clustering" => Ok(GeneratedField::EnableProximityClustering),
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
                formatter.write_str("struct virtengine.hpc.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut platform_fee_rate__ = None;
                let mut provider_reward_rate__ = None;
                let mut node_reward_rate__ = None;
                let mut min_job_duration_seconds__ = None;
                let mut max_job_duration_seconds__ = None;
                let mut default_identity_threshold__ = None;
                let mut cluster_heartbeat_timeout__ = None;
                let mut node_heartbeat_timeout__ = None;
                let mut latency_weight_factor__ = None;
                let mut capacity_weight_factor__ = None;
                let mut max_latency_ms__ = None;
                let mut dispute_resolution_period__ = None;
                let mut reward_formula_version__ = None;
                let mut enable_proximity_clustering__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::PlatformFeeRate => {
                            if platform_fee_rate__.is_some() {
                                return Err(serde::de::Error::duplicate_field("platformFeeRate"));
                            }
                            platform_fee_rate__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderRewardRate => {
                            if provider_reward_rate__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerRewardRate"));
                            }
                            provider_reward_rate__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NodeRewardRate => {
                            if node_reward_rate__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeRewardRate"));
                            }
                            node_reward_rate__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinJobDurationSeconds => {
                            if min_job_duration_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minJobDurationSeconds"));
                            }
                            min_job_duration_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxJobDurationSeconds => {
                            if max_job_duration_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxJobDurationSeconds"));
                            }
                            max_job_duration_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DefaultIdentityThreshold => {
                            if default_identity_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("defaultIdentityThreshold"));
                            }
                            default_identity_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ClusterHeartbeatTimeout => {
                            if cluster_heartbeat_timeout__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterHeartbeatTimeout"));
                            }
                            cluster_heartbeat_timeout__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NodeHeartbeatTimeout => {
                            if node_heartbeat_timeout__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeHeartbeatTimeout"));
                            }
                            node_heartbeat_timeout__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LatencyWeightFactor => {
                            if latency_weight_factor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latencyWeightFactor"));
                            }
                            latency_weight_factor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CapacityWeightFactor => {
                            if capacity_weight_factor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("capacityWeightFactor"));
                            }
                            capacity_weight_factor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxLatencyMs => {
                            if max_latency_ms__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxLatencyMs"));
                            }
                            max_latency_ms__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DisputeResolutionPeriod => {
                            if dispute_resolution_period__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeResolutionPeriod"));
                            }
                            dispute_resolution_period__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RewardFormulaVersion => {
                            if reward_formula_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardFormulaVersion"));
                            }
                            reward_formula_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EnableProximityClustering => {
                            if enable_proximity_clustering__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enableProximityClustering"));
                            }
                            enable_proximity_clustering__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Params {
                    platform_fee_rate: platform_fee_rate__.unwrap_or_default(),
                    provider_reward_rate: provider_reward_rate__.unwrap_or_default(),
                    node_reward_rate: node_reward_rate__.unwrap_or_default(),
                    min_job_duration_seconds: min_job_duration_seconds__.unwrap_or_default(),
                    max_job_duration_seconds: max_job_duration_seconds__.unwrap_or_default(),
                    default_identity_threshold: default_identity_threshold__.unwrap_or_default(),
                    cluster_heartbeat_timeout: cluster_heartbeat_timeout__.unwrap_or_default(),
                    node_heartbeat_timeout: node_heartbeat_timeout__.unwrap_or_default(),
                    latency_weight_factor: latency_weight_factor__.unwrap_or_default(),
                    capacity_weight_factor: capacity_weight_factor__.unwrap_or_default(),
                    max_latency_ms: max_latency_ms__.unwrap_or_default(),
                    dispute_resolution_period: dispute_resolution_period__.unwrap_or_default(),
                    reward_formula_version: reward_formula_version__.unwrap_or_default(),
                    enable_proximity_clustering: enable_proximity_clustering__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Partition {
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
        if self.nodes != 0 {
            len += 1;
        }
        if self.max_runtime != 0 {
            len += 1;
        }
        if self.default_runtime != 0 {
            len += 1;
        }
        if self.max_nodes != 0 {
            len += 1;
        }
        if !self.features.is_empty() {
            len += 1;
        }
        if self.priority != 0 {
            len += 1;
        }
        if !self.state.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.Partition", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if self.nodes != 0 {
            struct_ser.serialize_field("nodes", &self.nodes)?;
        }
        if self.max_runtime != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntime", ToString::to_string(&self.max_runtime).as_str())?;
        }
        if self.default_runtime != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("defaultRuntime", ToString::to_string(&self.default_runtime).as_str())?;
        }
        if self.max_nodes != 0 {
            struct_ser.serialize_field("maxNodes", &self.max_nodes)?;
        }
        if !self.features.is_empty() {
            struct_ser.serialize_field("features", &self.features)?;
        }
        if self.priority != 0 {
            struct_ser.serialize_field("priority", &self.priority)?;
        }
        if !self.state.is_empty() {
            struct_ser.serialize_field("state", &self.state)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Partition {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "nodes",
            "max_runtime",
            "maxRuntime",
            "default_runtime",
            "defaultRuntime",
            "max_nodes",
            "maxNodes",
            "features",
            "priority",
            "state",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Nodes,
            MaxRuntime,
            DefaultRuntime,
            MaxNodes,
            Features,
            Priority,
            State,
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
                            "nodes" => Ok(GeneratedField::Nodes),
                            "maxRuntime" | "max_runtime" => Ok(GeneratedField::MaxRuntime),
                            "defaultRuntime" | "default_runtime" => Ok(GeneratedField::DefaultRuntime),
                            "maxNodes" | "max_nodes" => Ok(GeneratedField::MaxNodes),
                            "features" => Ok(GeneratedField::Features),
                            "priority" => Ok(GeneratedField::Priority),
                            "state" => Ok(GeneratedField::State),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Partition;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.Partition")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Partition, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut nodes__ = None;
                let mut max_runtime__ = None;
                let mut default_runtime__ = None;
                let mut max_nodes__ = None;
                let mut features__ = None;
                let mut priority__ = None;
                let mut state__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Nodes => {
                            if nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodes"));
                            }
                            nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRuntime => {
                            if max_runtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntime"));
                            }
                            max_runtime__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DefaultRuntime => {
                            if default_runtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("defaultRuntime"));
                            }
                            default_runtime__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxNodes => {
                            if max_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxNodes"));
                            }
                            max_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Features => {
                            if features__.is_some() {
                                return Err(serde::de::Error::duplicate_field("features"));
                            }
                            features__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Priority => {
                            if priority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("priority"));
                            }
                            priority__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Partition {
                    name: name__.unwrap_or_default(),
                    nodes: nodes__.unwrap_or_default(),
                    max_runtime: max_runtime__.unwrap_or_default(),
                    default_runtime: default_runtime__.unwrap_or_default(),
                    max_nodes: max_nodes__.unwrap_or_default(),
                    features: features__.unwrap_or_default(),
                    priority: priority__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.Partition", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PreconfiguredWorkload {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.workload_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.container_image.is_empty() {
            len += 1;
        }
        if !self.default_command.is_empty() {
            len += 1;
        }
        if self.required_resources.is_some() {
            len += 1;
        }
        if !self.category.is_empty() {
            len += 1;
        }
        if !self.version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.PreconfiguredWorkload", len)?;
        if !self.workload_id.is_empty() {
            struct_ser.serialize_field("workloadId", &self.workload_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.container_image.is_empty() {
            struct_ser.serialize_field("containerImage", &self.container_image)?;
        }
        if !self.default_command.is_empty() {
            struct_ser.serialize_field("defaultCommand", &self.default_command)?;
        }
        if let Some(v) = self.required_resources.as_ref() {
            struct_ser.serialize_field("requiredResources", v)?;
        }
        if !self.category.is_empty() {
            struct_ser.serialize_field("category", &self.category)?;
        }
        if !self.version.is_empty() {
            struct_ser.serialize_field("version", &self.version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PreconfiguredWorkload {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "workload_id",
            "workloadId",
            "name",
            "description",
            "container_image",
            "containerImage",
            "default_command",
            "defaultCommand",
            "required_resources",
            "requiredResources",
            "category",
            "version",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            WorkloadId,
            Name,
            Description,
            ContainerImage,
            DefaultCommand,
            RequiredResources,
            Category,
            Version,
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
                            "workloadId" | "workload_id" => Ok(GeneratedField::WorkloadId),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "containerImage" | "container_image" => Ok(GeneratedField::ContainerImage),
                            "defaultCommand" | "default_command" => Ok(GeneratedField::DefaultCommand),
                            "requiredResources" | "required_resources" => Ok(GeneratedField::RequiredResources),
                            "category" => Ok(GeneratedField::Category),
                            "version" => Ok(GeneratedField::Version),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PreconfiguredWorkload;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.PreconfiguredWorkload")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PreconfiguredWorkload, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut workload_id__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut container_image__ = None;
                let mut default_command__ = None;
                let mut required_resources__ = None;
                let mut category__ = None;
                let mut version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::WorkloadId => {
                            if workload_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("workloadId"));
                            }
                            workload_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ContainerImage => {
                            if container_image__.is_some() {
                                return Err(serde::de::Error::duplicate_field("containerImage"));
                            }
                            container_image__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DefaultCommand => {
                            if default_command__.is_some() {
                                return Err(serde::de::Error::duplicate_field("defaultCommand"));
                            }
                            default_command__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequiredResources => {
                            if required_resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredResources"));
                            }
                            required_resources__ = map_.next_value()?;
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(PreconfiguredWorkload {
                    workload_id: workload_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    container_image: container_image__.unwrap_or_default(),
                    default_command: default_command__.unwrap_or_default(),
                    required_resources: required_resources__,
                    category: category__.unwrap_or_default(),
                    version: version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.PreconfiguredWorkload", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryClusterRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryClusterRequest", len)?;
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryClusterRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster_id",
            "clusterId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClusterId,
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
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryClusterRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryClusterRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryClusterRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryClusterRequest {
                    cluster_id: cluster_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryClusterRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryClusterResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryClusterResponse", len)?;
        if let Some(v) = self.cluster.as_ref() {
            struct_ser.serialize_field("cluster", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryClusterResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Cluster,
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
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryClusterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryClusterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryClusterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Cluster => {
                            if cluster__.is_some() {
                                return Err(serde::de::Error::duplicate_field("cluster"));
                            }
                            cluster__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryClusterResponse {
                    cluster: cluster__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryClusterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryClustersByProviderRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryClustersByProviderRequest", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryClustersByProviderRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
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
            type Value = QueryClustersByProviderRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryClustersByProviderRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryClustersByProviderRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryClustersByProviderRequest {
                    provider_address: provider_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryClustersByProviderRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryClustersByProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.clusters.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryClustersByProviderResponse", len)?;
        if !self.clusters.is_empty() {
            struct_ser.serialize_field("clusters", &self.clusters)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryClustersByProviderResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "clusters",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Clusters,
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
                            "clusters" => Ok(GeneratedField::Clusters),
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
            type Value = QueryClustersByProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryClustersByProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryClustersByProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut clusters__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Clusters => {
                            if clusters__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusters"));
                            }
                            clusters__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryClustersByProviderResponse {
                    clusters: clusters__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryClustersByProviderResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryClustersRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.state != 0 {
            len += 1;
        }
        if !self.region.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryClustersRequest", len)?;
        if self.state != 0 {
            let v = ClusterState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.region.is_empty() {
            struct_ser.serialize_field("region", &self.region)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryClustersRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "region",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            Region,
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
                            "state" => Ok(GeneratedField::State),
                            "region" => Ok(GeneratedField::Region),
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
            type Value = QueryClustersRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryClustersRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryClustersRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut region__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<ClusterState>()? as i32);
                        }
                        GeneratedField::Region => {
                            if region__.is_some() {
                                return Err(serde::de::Error::duplicate_field("region"));
                            }
                            region__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryClustersRequest {
                    state: state__.unwrap_or_default(),
                    region: region__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryClustersRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryClustersResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.clusters.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryClustersResponse", len)?;
        if !self.clusters.is_empty() {
            struct_ser.serialize_field("clusters", &self.clusters)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryClustersResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "clusters",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Clusters,
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
                            "clusters" => Ok(GeneratedField::Clusters),
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
            type Value = QueryClustersResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryClustersResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryClustersResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut clusters__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Clusters => {
                            if clusters__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusters"));
                            }
                            clusters__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryClustersResponse {
                    clusters: clusters__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryClustersResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDisputeRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.dispute_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryDisputeRequest", len)?;
        if !self.dispute_id.is_empty() {
            struct_ser.serialize_field("disputeId", &self.dispute_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDisputeRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "dispute_id",
            "disputeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DisputeId,
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
                            "disputeId" | "dispute_id" => Ok(GeneratedField::DisputeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDisputeRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryDisputeRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDisputeRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut dispute_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DisputeId => {
                            if dispute_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputeId"));
                            }
                            dispute_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDisputeRequest {
                    dispute_id: dispute_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryDisputeRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDisputeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.dispute.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryDisputeResponse", len)?;
        if let Some(v) = self.dispute.as_ref() {
            struct_ser.serialize_field("dispute", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDisputeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "dispute",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Dispute,
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
                            "dispute" => Ok(GeneratedField::Dispute),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDisputeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryDisputeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDisputeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut dispute__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Dispute => {
                            if dispute__.is_some() {
                                return Err(serde::de::Error::duplicate_field("dispute"));
                            }
                            dispute__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDisputeResponse {
                    dispute: dispute__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryDisputeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDisputesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.status != 0 {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryDisputesRequest", len)?;
        if self.status != 0 {
            let v = DisputeStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDisputesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "status",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Status,
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
                            "status" => Ok(GeneratedField::Status),
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
            type Value = QueryDisputesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryDisputesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDisputesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut status__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<DisputeStatus>()? as i32);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDisputesRequest {
                    status: status__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryDisputesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDisputesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.disputes.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryDisputesResponse", len)?;
        if !self.disputes.is_empty() {
            struct_ser.serialize_field("disputes", &self.disputes)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDisputesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "disputes",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Disputes,
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
                            "disputes" => Ok(GeneratedField::Disputes),
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
            type Value = QueryDisputesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryDisputesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDisputesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut disputes__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Disputes => {
                            if disputes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("disputes"));
                            }
                            disputes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDisputesResponse {
                    disputes: disputes__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryDisputesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobAccountingRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobAccountingRequest", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobAccountingRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryJobAccountingRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobAccountingRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobAccountingRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryJobAccountingRequest {
                    job_id: job_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobAccountingRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobAccountingResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.accounting.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobAccountingResponse", len)?;
        if let Some(v) = self.accounting.as_ref() {
            struct_ser.serialize_field("accounting", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobAccountingResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "accounting",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Accounting,
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
                            "accounting" => Ok(GeneratedField::Accounting),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryJobAccountingResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobAccountingResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobAccountingResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut accounting__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Accounting => {
                            if accounting__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accounting"));
                            }
                            accounting__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobAccountingResponse {
                    accounting: accounting__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobAccountingResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobRequest", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryJobRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryJobRequest {
                    job_id: job_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.job.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobResponse", len)?;
        if let Some(v) = self.job.as_ref() {
            struct_ser.serialize_field("job", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Job,
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
                            "job" => Ok(GeneratedField::Job),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryJobResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Job => {
                            if job__.is_some() {
                                return Err(serde::de::Error::duplicate_field("job"));
                            }
                            job__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobResponse {
                    job: job__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobsByCustomerRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.customer_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobsByCustomerRequest", len)?;
        if !self.customer_address.is_empty() {
            struct_ser.serialize_field("customerAddress", &self.customer_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobsByCustomerRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "customer_address",
            "customerAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CustomerAddress,
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
                            "customerAddress" | "customer_address" => Ok(GeneratedField::CustomerAddress),
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
            type Value = QueryJobsByCustomerRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobsByCustomerRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobsByCustomerRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut customer_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CustomerAddress => {
                            if customer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customerAddress"));
                            }
                            customer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobsByCustomerRequest {
                    customer_address: customer_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobsByCustomerRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobsByCustomerResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.jobs.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobsByCustomerResponse", len)?;
        if !self.jobs.is_empty() {
            struct_ser.serialize_field("jobs", &self.jobs)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobsByCustomerResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "jobs",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Jobs,
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
                            "jobs" => Ok(GeneratedField::Jobs),
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
            type Value = QueryJobsByCustomerResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobsByCustomerResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobsByCustomerResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut jobs__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Jobs => {
                            if jobs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobs"));
                            }
                            jobs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobsByCustomerResponse {
                    jobs: jobs__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobsByCustomerResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobsByProviderRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobsByProviderRequest", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobsByProviderRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
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
            type Value = QueryJobsByProviderRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobsByProviderRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobsByProviderRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobsByProviderRequest {
                    provider_address: provider_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobsByProviderRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobsByProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.jobs.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobsByProviderResponse", len)?;
        if !self.jobs.is_empty() {
            struct_ser.serialize_field("jobs", &self.jobs)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobsByProviderResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "jobs",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Jobs,
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
                            "jobs" => Ok(GeneratedField::Jobs),
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
            type Value = QueryJobsByProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobsByProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobsByProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut jobs__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Jobs => {
                            if jobs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobs"));
                            }
                            jobs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobsByProviderResponse {
                    jobs: jobs__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobsByProviderResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.state != 0 {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobsRequest", len)?;
        if self.state != 0 {
            let v = JobState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "cluster_id",
            "clusterId",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            ClusterId,
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
                            "state" => Ok(GeneratedField::State),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
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
            type Value = QueryJobsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut cluster_id__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<JobState>()? as i32);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobsRequest {
                    state: state__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryJobsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.jobs.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryJobsResponse", len)?;
        if !self.jobs.is_empty() {
            struct_ser.serialize_field("jobs", &self.jobs)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryJobsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "jobs",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Jobs,
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
                            "jobs" => Ok(GeneratedField::Jobs),
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
            type Value = QueryJobsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryJobsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryJobsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut jobs__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Jobs => {
                            if jobs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobs"));
                            }
                            jobs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryJobsResponse {
                    jobs: jobs__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryJobsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryNodeMetadataRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.node_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryNodeMetadataRequest", len)?;
        if !self.node_id.is_empty() {
            struct_ser.serialize_field("nodeId", &self.node_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryNodeMetadataRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "node_id",
            "nodeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            NodeId,
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
                            "nodeId" | "node_id" => Ok(GeneratedField::NodeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryNodeMetadataRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryNodeMetadataRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryNodeMetadataRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut node_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::NodeId => {
                            if node_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeId"));
                            }
                            node_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryNodeMetadataRequest {
                    node_id: node_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryNodeMetadataRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryNodeMetadataResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.node.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryNodeMetadataResponse", len)?;
        if let Some(v) = self.node.as_ref() {
            struct_ser.serialize_field("node", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryNodeMetadataResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "node",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Node,
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
                            "node" => Ok(GeneratedField::Node),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryNodeMetadataResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryNodeMetadataResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryNodeMetadataResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut node__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Node => {
                            if node__.is_some() {
                                return Err(serde::de::Error::duplicate_field("node"));
                            }
                            node__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryNodeMetadataResponse {
                    node: node__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryNodeMetadataResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryNodesByClusterRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryNodesByClusterRequest", len)?;
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryNodesByClusterRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster_id",
            "clusterId",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClusterId,
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
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
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
            type Value = QueryNodesByClusterRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryNodesByClusterRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryNodesByClusterRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster_id__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryNodesByClusterRequest {
                    cluster_id: cluster_id__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryNodesByClusterRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryNodesByClusterResponse {
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
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryNodesByClusterResponse", len)?;
        if !self.nodes.is_empty() {
            struct_ser.serialize_field("nodes", &self.nodes)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryNodesByClusterResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "nodes",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Nodes,
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
                            "nodes" => Ok(GeneratedField::Nodes),
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
            type Value = QueryNodesByClusterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryNodesByClusterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryNodesByClusterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut nodes__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Nodes => {
                            if nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodes"));
                            }
                            nodes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryNodesByClusterResponse {
                    nodes: nodes__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryNodesByClusterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.offering_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryOfferingRequest", len)?;
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offering_id",
            "offeringId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            OfferingId,
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
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryOfferingRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryOfferingRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offering_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryOfferingRequest {
                    offering_id: offering_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryOfferingRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.offering.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryOfferingResponse", len)?;
        if let Some(v) = self.offering.as_ref() {
            struct_ser.serialize_field("offering", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offering",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Offering,
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
                            "offering" => Ok(GeneratedField::Offering),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryOfferingResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryOfferingResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offering__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Offering => {
                            if offering__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offering"));
                            }
                            offering__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryOfferingResponse {
                    offering: offering__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryOfferingResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingsByClusterRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryOfferingsByClusterRequest", len)?;
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingsByClusterRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "cluster_id",
            "clusterId",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClusterId,
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
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
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
            type Value = QueryOfferingsByClusterRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryOfferingsByClusterRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingsByClusterRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut cluster_id__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryOfferingsByClusterRequest {
                    cluster_id: cluster_id__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryOfferingsByClusterRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingsByClusterResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.offerings.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryOfferingsByClusterResponse", len)?;
        if !self.offerings.is_empty() {
            struct_ser.serialize_field("offerings", &self.offerings)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingsByClusterResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offerings",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Offerings,
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
                            "offerings" => Ok(GeneratedField::Offerings),
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
            type Value = QueryOfferingsByClusterResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryOfferingsByClusterResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingsByClusterResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offerings__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Offerings => {
                            if offerings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offerings"));
                            }
                            offerings__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryOfferingsByClusterResponse {
                    offerings: offerings__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryOfferingsByClusterResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.active_only {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryOfferingsRequest", len)?;
        if self.active_only {
            struct_ser.serialize_field("activeOnly", &self.active_only)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "active_only",
            "activeOnly",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ActiveOnly,
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
                            "activeOnly" | "active_only" => Ok(GeneratedField::ActiveOnly),
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
            type Value = QueryOfferingsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryOfferingsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut active_only__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ActiveOnly => {
                            if active_only__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activeOnly"));
                            }
                            active_only__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryOfferingsRequest {
                    active_only: active_only__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryOfferingsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.offerings.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryOfferingsResponse", len)?;
        if !self.offerings.is_empty() {
            struct_ser.serialize_field("offerings", &self.offerings)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offerings",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Offerings,
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
                            "offerings" => Ok(GeneratedField::Offerings),
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
            type Value = QueryOfferingsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryOfferingsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offerings__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Offerings => {
                            if offerings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offerings"));
                            }
                            offerings__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryOfferingsResponse {
                    offerings: offerings__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryOfferingsResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.hpc.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.hpc.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRewardRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reward_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryRewardRequest", len)?;
        if !self.reward_id.is_empty() {
            struct_ser.serialize_field("rewardId", &self.reward_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRewardRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reward_id",
            "rewardId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RewardId,
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
                            "rewardId" | "reward_id" => Ok(GeneratedField::RewardId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryRewardRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryRewardRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRewardRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reward_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RewardId => {
                            if reward_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardId"));
                            }
                            reward_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryRewardRequest {
                    reward_id: reward_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryRewardRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRewardResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.reward.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryRewardResponse", len)?;
        if let Some(v) = self.reward.as_ref() {
            struct_ser.serialize_field("reward", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRewardResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reward",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reward,
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
                            "reward" => Ok(GeneratedField::Reward),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryRewardResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryRewardResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRewardResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reward__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reward => {
                            if reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reward"));
                            }
                            reward__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryRewardResponse {
                    reward: reward__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryRewardResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRewardsByJobRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryRewardsByJobRequest", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRewardsByJobRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryRewardsByJobRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryRewardsByJobRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRewardsByJobRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryRewardsByJobRequest {
                    job_id: job_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryRewardsByJobRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRewardsByJobResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.rewards.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueryRewardsByJobResponse", len)?;
        if !self.rewards.is_empty() {
            struct_ser.serialize_field("rewards", &self.rewards)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRewardsByJobResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "rewards",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Rewards,
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
                            "rewards" => Ok(GeneratedField::Rewards),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryRewardsByJobResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueryRewardsByJobResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRewardsByJobResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut rewards__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Rewards => {
                            if rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewards"));
                            }
                            rewards__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryRewardsByJobResponse {
                    rewards: rewards__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueryRewardsByJobResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuerySchedulingDecisionByJobRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.job_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionByJobRequest", len)?;
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuerySchedulingDecisionByJobRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "job_id",
            "jobId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            JobId,
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
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuerySchedulingDecisionByJobRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QuerySchedulingDecisionByJobRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuerySchedulingDecisionByJobRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut job_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QuerySchedulingDecisionByJobRequest {
                    job_id: job_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionByJobRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuerySchedulingDecisionByJobResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.decision.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionByJobResponse", len)?;
        if let Some(v) = self.decision.as_ref() {
            struct_ser.serialize_field("decision", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuerySchedulingDecisionByJobResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "decision",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Decision,
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
                            "decision" => Ok(GeneratedField::Decision),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuerySchedulingDecisionByJobResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QuerySchedulingDecisionByJobResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuerySchedulingDecisionByJobResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut decision__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Decision => {
                            if decision__.is_some() {
                                return Err(serde::de::Error::duplicate_field("decision"));
                            }
                            decision__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QuerySchedulingDecisionByJobResponse {
                    decision: decision__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionByJobResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuerySchedulingDecisionRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.decision_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionRequest", len)?;
        if !self.decision_id.is_empty() {
            struct_ser.serialize_field("decisionId", &self.decision_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuerySchedulingDecisionRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "decision_id",
            "decisionId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DecisionId,
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
                            "decisionId" | "decision_id" => Ok(GeneratedField::DecisionId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuerySchedulingDecisionRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QuerySchedulingDecisionRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuerySchedulingDecisionRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut decision_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DecisionId => {
                            if decision_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("decisionId"));
                            }
                            decision_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QuerySchedulingDecisionRequest {
                    decision_id: decision_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QuerySchedulingDecisionResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.decision.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionResponse", len)?;
        if let Some(v) = self.decision.as_ref() {
            struct_ser.serialize_field("decision", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QuerySchedulingDecisionResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "decision",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Decision,
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
                            "decision" => Ok(GeneratedField::Decision),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QuerySchedulingDecisionResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QuerySchedulingDecisionResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QuerySchedulingDecisionResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut decision__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Decision => {
                            if decision__.is_some() {
                                return Err(serde::de::Error::duplicate_field("decision"));
                            }
                            decision__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QuerySchedulingDecisionResponse {
                    decision: decision__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QuerySchedulingDecisionResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueueOption {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.partition_name.is_empty() {
            len += 1;
        }
        if !self.display_name.is_empty() {
            len += 1;
        }
        if self.max_nodes != 0 {
            len += 1;
        }
        if self.max_runtime != 0 {
            len += 1;
        }
        if !self.features.is_empty() {
            len += 1;
        }
        if !self.price_multiplier.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.QueueOption", len)?;
        if !self.partition_name.is_empty() {
            struct_ser.serialize_field("partitionName", &self.partition_name)?;
        }
        if !self.display_name.is_empty() {
            struct_ser.serialize_field("displayName", &self.display_name)?;
        }
        if self.max_nodes != 0 {
            struct_ser.serialize_field("maxNodes", &self.max_nodes)?;
        }
        if self.max_runtime != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRuntime", ToString::to_string(&self.max_runtime).as_str())?;
        }
        if !self.features.is_empty() {
            struct_ser.serialize_field("features", &self.features)?;
        }
        if !self.price_multiplier.is_empty() {
            struct_ser.serialize_field("priceMultiplier", &self.price_multiplier)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueueOption {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "partition_name",
            "partitionName",
            "display_name",
            "displayName",
            "max_nodes",
            "maxNodes",
            "max_runtime",
            "maxRuntime",
            "features",
            "price_multiplier",
            "priceMultiplier",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            PartitionName,
            DisplayName,
            MaxNodes,
            MaxRuntime,
            Features,
            PriceMultiplier,
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
                            "partitionName" | "partition_name" => Ok(GeneratedField::PartitionName),
                            "displayName" | "display_name" => Ok(GeneratedField::DisplayName),
                            "maxNodes" | "max_nodes" => Ok(GeneratedField::MaxNodes),
                            "maxRuntime" | "max_runtime" => Ok(GeneratedField::MaxRuntime),
                            "features" => Ok(GeneratedField::Features),
                            "priceMultiplier" | "price_multiplier" => Ok(GeneratedField::PriceMultiplier),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueueOption;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.QueueOption")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueueOption, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut partition_name__ = None;
                let mut display_name__ = None;
                let mut max_nodes__ = None;
                let mut max_runtime__ = None;
                let mut features__ = None;
                let mut price_multiplier__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::PartitionName => {
                            if partition_name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("partitionName"));
                            }
                            partition_name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DisplayName => {
                            if display_name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("displayName"));
                            }
                            display_name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxNodes => {
                            if max_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxNodes"));
                            }
                            max_nodes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRuntime => {
                            if max_runtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRuntime"));
                            }
                            max_runtime__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Features => {
                            if features__.is_some() {
                                return Err(serde::de::Error::duplicate_field("features"));
                            }
                            features__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PriceMultiplier => {
                            if price_multiplier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("priceMultiplier"));
                            }
                            price_multiplier__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueueOption {
                    partition_name: partition_name__.unwrap_or_default(),
                    display_name: display_name__.unwrap_or_default(),
                    max_nodes: max_nodes__.unwrap_or_default(),
                    max_runtime: max_runtime__.unwrap_or_default(),
                    features: features__.unwrap_or_default(),
                    price_multiplier: price_multiplier__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.QueueOption", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RewardCalculationDetails {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.total_usage_value.is_empty() {
            len += 1;
        }
        if !self.reward_pool_contribution.is_empty() {
            len += 1;
        }
        if !self.platform_fee_rate.is_empty() {
            len += 1;
        }
        if !self.node_contribution_formula.is_empty() {
            len += 1;
        }
        if !self.input_metrics.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.RewardCalculationDetails", len)?;
        if !self.total_usage_value.is_empty() {
            struct_ser.serialize_field("totalUsageValue", &self.total_usage_value)?;
        }
        if !self.reward_pool_contribution.is_empty() {
            struct_ser.serialize_field("rewardPoolContribution", &self.reward_pool_contribution)?;
        }
        if !self.platform_fee_rate.is_empty() {
            struct_ser.serialize_field("platformFeeRate", &self.platform_fee_rate)?;
        }
        if !self.node_contribution_formula.is_empty() {
            struct_ser.serialize_field("nodeContributionFormula", &self.node_contribution_formula)?;
        }
        if !self.input_metrics.is_empty() {
            struct_ser.serialize_field("inputMetrics", &self.input_metrics)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RewardCalculationDetails {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "total_usage_value",
            "totalUsageValue",
            "reward_pool_contribution",
            "rewardPoolContribution",
            "platform_fee_rate",
            "platformFeeRate",
            "node_contribution_formula",
            "nodeContributionFormula",
            "input_metrics",
            "inputMetrics",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TotalUsageValue,
            RewardPoolContribution,
            PlatformFeeRate,
            NodeContributionFormula,
            InputMetrics,
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
                            "totalUsageValue" | "total_usage_value" => Ok(GeneratedField::TotalUsageValue),
                            "rewardPoolContribution" | "reward_pool_contribution" => Ok(GeneratedField::RewardPoolContribution),
                            "platformFeeRate" | "platform_fee_rate" => Ok(GeneratedField::PlatformFeeRate),
                            "nodeContributionFormula" | "node_contribution_formula" => Ok(GeneratedField::NodeContributionFormula),
                            "inputMetrics" | "input_metrics" => Ok(GeneratedField::InputMetrics),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RewardCalculationDetails;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.RewardCalculationDetails")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RewardCalculationDetails, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut total_usage_value__ = None;
                let mut reward_pool_contribution__ = None;
                let mut platform_fee_rate__ = None;
                let mut node_contribution_formula__ = None;
                let mut input_metrics__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TotalUsageValue => {
                            if total_usage_value__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalUsageValue"));
                            }
                            total_usage_value__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RewardPoolContribution => {
                            if reward_pool_contribution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardPoolContribution"));
                            }
                            reward_pool_contribution__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PlatformFeeRate => {
                            if platform_fee_rate__.is_some() {
                                return Err(serde::de::Error::duplicate_field("platformFeeRate"));
                            }
                            platform_fee_rate__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NodeContributionFormula => {
                            if node_contribution_formula__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nodeContributionFormula"));
                            }
                            node_contribution_formula__ = Some(map_.next_value()?);
                        }
                        GeneratedField::InputMetrics => {
                            if input_metrics__.is_some() {
                                return Err(serde::de::Error::duplicate_field("inputMetrics"));
                            }
                            input_metrics__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(RewardCalculationDetails {
                    total_usage_value: total_usage_value__.unwrap_or_default(),
                    reward_pool_contribution: reward_pool_contribution__.unwrap_or_default(),
                    platform_fee_rate: platform_fee_rate__.unwrap_or_default(),
                    node_contribution_formula: node_contribution_formula__.unwrap_or_default(),
                    input_metrics: input_metrics__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.RewardCalculationDetails", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SchedulingDecision {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.decision_id.is_empty() {
            len += 1;
        }
        if !self.job_id.is_empty() {
            len += 1;
        }
        if !self.selected_cluster_id.is_empty() {
            len += 1;
        }
        if !self.candidate_clusters.is_empty() {
            len += 1;
        }
        if !self.decision_reason.is_empty() {
            len += 1;
        }
        if self.is_fallback {
            len += 1;
        }
        if !self.fallback_reason.is_empty() {
            len += 1;
        }
        if !self.latency_score.is_empty() {
            len += 1;
        }
        if !self.capacity_score.is_empty() {
            len += 1;
        }
        if !self.combined_score.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.hpc.v1.SchedulingDecision", len)?;
        if !self.decision_id.is_empty() {
            struct_ser.serialize_field("decisionId", &self.decision_id)?;
        }
        if !self.job_id.is_empty() {
            struct_ser.serialize_field("jobId", &self.job_id)?;
        }
        if !self.selected_cluster_id.is_empty() {
            struct_ser.serialize_field("selectedClusterId", &self.selected_cluster_id)?;
        }
        if !self.candidate_clusters.is_empty() {
            struct_ser.serialize_field("candidateClusters", &self.candidate_clusters)?;
        }
        if !self.decision_reason.is_empty() {
            struct_ser.serialize_field("decisionReason", &self.decision_reason)?;
        }
        if self.is_fallback {
            struct_ser.serialize_field("isFallback", &self.is_fallback)?;
        }
        if !self.fallback_reason.is_empty() {
            struct_ser.serialize_field("fallbackReason", &self.fallback_reason)?;
        }
        if !self.latency_score.is_empty() {
            struct_ser.serialize_field("latencyScore", &self.latency_score)?;
        }
        if !self.capacity_score.is_empty() {
            struct_ser.serialize_field("capacityScore", &self.capacity_score)?;
        }
        if !self.combined_score.is_empty() {
            struct_ser.serialize_field("combinedScore", &self.combined_score)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SchedulingDecision {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "decision_id",
            "decisionId",
            "job_id",
            "jobId",
            "selected_cluster_id",
            "selectedClusterId",
            "candidate_clusters",
            "candidateClusters",
            "decision_reason",
            "decisionReason",
            "is_fallback",
            "isFallback",
            "fallback_reason",
            "fallbackReason",
            "latency_score",
            "latencyScore",
            "capacity_score",
            "capacityScore",
            "combined_score",
            "combinedScore",
            "created_at",
            "createdAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DecisionId,
            JobId,
            SelectedClusterId,
            CandidateClusters,
            DecisionReason,
            IsFallback,
            FallbackReason,
            LatencyScore,
            CapacityScore,
            CombinedScore,
            CreatedAt,
            BlockHeight,
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
                            "decisionId" | "decision_id" => Ok(GeneratedField::DecisionId),
                            "jobId" | "job_id" => Ok(GeneratedField::JobId),
                            "selectedClusterId" | "selected_cluster_id" => Ok(GeneratedField::SelectedClusterId),
                            "candidateClusters" | "candidate_clusters" => Ok(GeneratedField::CandidateClusters),
                            "decisionReason" | "decision_reason" => Ok(GeneratedField::DecisionReason),
                            "isFallback" | "is_fallback" => Ok(GeneratedField::IsFallback),
                            "fallbackReason" | "fallback_reason" => Ok(GeneratedField::FallbackReason),
                            "latencyScore" | "latency_score" => Ok(GeneratedField::LatencyScore),
                            "capacityScore" | "capacity_score" => Ok(GeneratedField::CapacityScore),
                            "combinedScore" | "combined_score" => Ok(GeneratedField::CombinedScore),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SchedulingDecision;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.hpc.v1.SchedulingDecision")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SchedulingDecision, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut decision_id__ = None;
                let mut job_id__ = None;
                let mut selected_cluster_id__ = None;
                let mut candidate_clusters__ = None;
                let mut decision_reason__ = None;
                let mut is_fallback__ = None;
                let mut fallback_reason__ = None;
                let mut latency_score__ = None;
                let mut capacity_score__ = None;
                let mut combined_score__ = None;
                let mut created_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DecisionId => {
                            if decision_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("decisionId"));
                            }
                            decision_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JobId => {
                            if job_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jobId"));
                            }
                            job_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SelectedClusterId => {
                            if selected_cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("selectedClusterId"));
                            }
                            selected_cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CandidateClusters => {
                            if candidate_clusters__.is_some() {
                                return Err(serde::de::Error::duplicate_field("candidateClusters"));
                            }
                            candidate_clusters__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DecisionReason => {
                            if decision_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("decisionReason"));
                            }
                            decision_reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsFallback => {
                            if is_fallback__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isFallback"));
                            }
                            is_fallback__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FallbackReason => {
                            if fallback_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fallbackReason"));
                            }
                            fallback_reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LatencyScore => {
                            if latency_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latencyScore"));
                            }
                            latency_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CapacityScore => {
                            if capacity_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("capacityScore"));
                            }
                            capacity_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CombinedScore => {
                            if combined_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("combinedScore"));
                            }
                            combined_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(SchedulingDecision {
                    decision_id: decision_id__.unwrap_or_default(),
                    job_id: job_id__.unwrap_or_default(),
                    selected_cluster_id: selected_cluster_id__.unwrap_or_default(),
                    candidate_clusters: candidate_clusters__.unwrap_or_default(),
                    decision_reason: decision_reason__.unwrap_or_default(),
                    is_fallback: is_fallback__.unwrap_or_default(),
                    fallback_reason: fallback_reason__.unwrap_or_default(),
                    latency_score: latency_score__.unwrap_or_default(),
                    capacity_score: capacity_score__.unwrap_or_default(),
                    combined_score: combined_score__.unwrap_or_default(),
                    created_at: created_at__,
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.hpc.v1.SchedulingDecision", FIELDS, GeneratedVisitor)
    }
}

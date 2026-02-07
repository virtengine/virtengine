// @generated
impl serde::Serialize for AnomalyDetectedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.anomaly_type.is_empty() {
            len += 1;
        }
        if !self.severity.is_empty() {
            len += 1;
        }
        if self.detected_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.AnomalyDetectedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.anomaly_type.is_empty() {
            struct_ser.serialize_field("anomalyType", &self.anomaly_type)?;
        }
        if !self.severity.is_empty() {
            struct_ser.serialize_field("severity", &self.severity)?;
        }
        if self.detected_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("detectedAt", ToString::to_string(&self.detected_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AnomalyDetectedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "anomaly_type",
            "anomalyType",
            "severity",
            "detected_at",
            "detectedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            AnomalyType,
            Severity,
            DetectedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "anomalyType" | "anomaly_type" => Ok(GeneratedField::AnomalyType),
                            "severity" => Ok(GeneratedField::Severity),
                            "detectedAt" | "detected_at" => Ok(GeneratedField::DetectedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AnomalyDetectedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.AnomalyDetectedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AnomalyDetectedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut anomaly_type__ = None;
                let mut severity__ = None;
                let mut detected_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AnomalyType => {
                            if anomaly_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("anomalyType"));
                            }
                            anomaly_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Severity => {
                            if severity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("severity"));
                            }
                            severity__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DetectedAt => {
                            if detected_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("detectedAt"));
                            }
                            detected_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AnomalyDetectedEvent {
                    provider: provider__.unwrap_or_default(),
                    anomaly_type: anomaly_type__.unwrap_or_default(),
                    severity: severity__.unwrap_or_default(),
                    detected_at: detected_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.AnomalyDetectedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AnomalyFlag {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.reporter.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        if !self.status.is_empty() {
            len += 1;
        }
        if self.flagged_at != 0 {
            len += 1;
        }
        if !self.resolution.is_empty() {
            len += 1;
        }
        if self.resolved_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.AnomalyFlag", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", &self.reporter)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        if !self.status.is_empty() {
            struct_ser.serialize_field("status", &self.status)?;
        }
        if self.flagged_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("flaggedAt", ToString::to_string(&self.flagged_at).as_str())?;
        }
        if !self.resolution.is_empty() {
            struct_ser.serialize_field("resolution", &self.resolution)?;
        }
        if self.resolved_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("resolvedAt", ToString::to_string(&self.resolved_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AnomalyFlag {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "reporter",
            "reason",
            "evidence",
            "status",
            "flagged_at",
            "flaggedAt",
            "resolution",
            "resolved_at",
            "resolvedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            Reporter,
            Reason,
            Evidence,
            Status,
            FlaggedAt,
            Resolution,
            ResolvedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "reporter" => Ok(GeneratedField::Reporter),
                            "reason" => Ok(GeneratedField::Reason),
                            "evidence" => Ok(GeneratedField::Evidence),
                            "status" => Ok(GeneratedField::Status),
                            "flaggedAt" | "flagged_at" => Ok(GeneratedField::FlaggedAt),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "resolvedAt" | "resolved_at" => Ok(GeneratedField::ResolvedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AnomalyFlag;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.AnomalyFlag")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AnomalyFlag, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut reporter__ = None;
                let mut reason__ = None;
                let mut evidence__ = None;
                let mut status__ = None;
                let mut flagged_at__ = None;
                let mut resolution__ = None;
                let mut resolved_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = Some(map_.next_value()?);
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
                            status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FlaggedAt => {
                            if flagged_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("flaggedAt"));
                            }
                            flagged_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResolvedAt => {
                            if resolved_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolvedAt"));
                            }
                            resolved_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AnomalyFlag {
                    provider: provider__.unwrap_or_default(),
                    reporter: reporter__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    flagged_at: flagged_at__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    resolved_at: resolved_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.AnomalyFlag", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AnomalyResolvedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.resolution.is_empty() {
            len += 1;
        }
        if self.resolved_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.AnomalyResolvedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.resolution.is_empty() {
            struct_ser.serialize_field("resolution", &self.resolution)?;
        }
        if self.resolved_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("resolvedAt", ToString::to_string(&self.resolved_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AnomalyResolvedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "resolution",
            "resolved_at",
            "resolvedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            Resolution,
            ResolvedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "resolvedAt" | "resolved_at" => Ok(GeneratedField::ResolvedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AnomalyResolvedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.AnomalyResolvedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AnomalyResolvedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut resolution__ = None;
                let mut resolved_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResolvedAt => {
                            if resolved_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolvedAt"));
                            }
                            resolved_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AnomalyResolvedEvent {
                    provider: provider__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    resolved_at: resolved_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.AnomalyResolvedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for BenchmarkResult {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.benchmark_type.is_empty() {
            len += 1;
        }
        if !self.score.is_empty() {
            len += 1;
        }
        if self.timestamp != 0 {
            len += 1;
        }
        if !self.hardware_info.is_empty() {
            len += 1;
        }
        if !self.raw_data.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.BenchmarkResult", len)?;
        if !self.benchmark_type.is_empty() {
            struct_ser.serialize_field("benchmarkType", &self.benchmark_type)?;
        }
        if !self.score.is_empty() {
            struct_ser.serialize_field("score", &self.score)?;
        }
        if self.timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("timestamp", ToString::to_string(&self.timestamp).as_str())?;
        }
        if !self.hardware_info.is_empty() {
            struct_ser.serialize_field("hardwareInfo", &self.hardware_info)?;
        }
        if !self.raw_data.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("rawData", pbjson::private::base64::encode(&self.raw_data).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BenchmarkResult {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "benchmark_type",
            "benchmarkType",
            "score",
            "timestamp",
            "hardware_info",
            "hardwareInfo",
            "raw_data",
            "rawData",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            BenchmarkType,
            Score,
            Timestamp,
            HardwareInfo,
            RawData,
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
                            "benchmarkType" | "benchmark_type" => Ok(GeneratedField::BenchmarkType),
                            "score" => Ok(GeneratedField::Score),
                            "timestamp" => Ok(GeneratedField::Timestamp),
                            "hardwareInfo" | "hardware_info" => Ok(GeneratedField::HardwareInfo),
                            "rawData" | "raw_data" => Ok(GeneratedField::RawData),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BenchmarkResult;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.BenchmarkResult")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<BenchmarkResult, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut benchmark_type__ = None;
                let mut score__ = None;
                let mut timestamp__ = None;
                let mut hardware_info__ = None;
                let mut raw_data__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::BenchmarkType => {
                            if benchmark_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("benchmarkType"));
                            }
                            benchmark_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::HardwareInfo => {
                            if hardware_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hardwareInfo"));
                            }
                            hardware_info__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RawData => {
                            if raw_data__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rawData"));
                            }
                            raw_data__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(BenchmarkResult {
                    benchmark_type: benchmark_type__.unwrap_or_default(),
                    score: score__.unwrap_or_default(),
                    timestamp: timestamp__.unwrap_or_default(),
                    hardware_info: hardware_info__.unwrap_or_default(),
                    raw_data: raw_data__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.BenchmarkResult", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for BenchmarksPrunedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if self.pruned_count != 0 {
            len += 1;
        }
        if self.pruned_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.BenchmarksPrunedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if self.pruned_count != 0 {
            struct_ser.serialize_field("prunedCount", &self.pruned_count)?;
        }
        if self.pruned_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("prunedAt", ToString::to_string(&self.pruned_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BenchmarksPrunedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "pruned_count",
            "prunedCount",
            "pruned_at",
            "prunedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            PrunedCount,
            PrunedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "prunedCount" | "pruned_count" => Ok(GeneratedField::PrunedCount),
                            "prunedAt" | "pruned_at" => Ok(GeneratedField::PrunedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BenchmarksPrunedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.BenchmarksPrunedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<BenchmarksPrunedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut pruned_count__ = None;
                let mut pruned_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PrunedCount => {
                            if pruned_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("prunedCount"));
                            }
                            pruned_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PrunedAt => {
                            if pruned_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("prunedAt"));
                            }
                            pruned_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(BenchmarksPrunedEvent {
                    provider: provider__.unwrap_or_default(),
                    pruned_count: pruned_count__.unwrap_or_default(),
                    pruned_at: pruned_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.BenchmarksPrunedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for BenchmarksSubmittedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if self.result_count != 0 {
            len += 1;
        }
        if self.submitted_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.BenchmarksSubmittedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if self.result_count != 0 {
            struct_ser.serialize_field("resultCount", &self.result_count)?;
        }
        if self.submitted_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("submittedAt", ToString::to_string(&self.submitted_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BenchmarksSubmittedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "cluster_id",
            "clusterId",
            "result_count",
            "resultCount",
            "submitted_at",
            "submittedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            ClusterId,
            ResultCount,
            SubmittedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "resultCount" | "result_count" => Ok(GeneratedField::ResultCount),
                            "submittedAt" | "submitted_at" => Ok(GeneratedField::SubmittedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BenchmarksSubmittedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.BenchmarksSubmittedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<BenchmarksSubmittedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut cluster_id__ = None;
                let mut result_count__ = None;
                let mut submitted_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResultCount => {
                            if result_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resultCount"));
                            }
                            result_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SubmittedAt => {
                            if submitted_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submittedAt"));
                            }
                            submitted_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(BenchmarksSubmittedEvent {
                    provider: provider__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    result_count: result_count__.unwrap_or_default(),
                    submitted_at: submitted_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.BenchmarksSubmittedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Challenge {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.requester.is_empty() {
            len += 1;
        }
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.benchmark_type.is_empty() {
            len += 1;
        }
        if !self.status.is_empty() {
            len += 1;
        }
        if self.requested_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if self.response.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.Challenge", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.requester.is_empty() {
            struct_ser.serialize_field("requester", &self.requester)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.benchmark_type.is_empty() {
            struct_ser.serialize_field("benchmarkType", &self.benchmark_type)?;
        }
        if !self.status.is_empty() {
            struct_ser.serialize_field("status", &self.status)?;
        }
        if self.requested_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("requestedAt", ToString::to_string(&self.requested_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if let Some(v) = self.response.as_ref() {
            struct_ser.serialize_field("response", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Challenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "requester",
            "provider",
            "benchmark_type",
            "benchmarkType",
            "status",
            "requested_at",
            "requestedAt",
            "expires_at",
            "expiresAt",
            "response",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            Requester,
            Provider,
            BenchmarkType,
            Status,
            RequestedAt,
            ExpiresAt,
            Response,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "requester" => Ok(GeneratedField::Requester),
                            "provider" => Ok(GeneratedField::Provider),
                            "benchmarkType" | "benchmark_type" => Ok(GeneratedField::BenchmarkType),
                            "status" => Ok(GeneratedField::Status),
                            "requestedAt" | "requested_at" => Ok(GeneratedField::RequestedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "response" => Ok(GeneratedField::Response),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Challenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.Challenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Challenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut requester__ = None;
                let mut provider__ = None;
                let mut benchmark_type__ = None;
                let mut status__ = None;
                let mut requested_at__ = None;
                let mut expires_at__ = None;
                let mut response__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Requester => {
                            if requester__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requester"));
                            }
                            requester__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BenchmarkType => {
                            if benchmark_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("benchmarkType"));
                            }
                            benchmark_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequestedAt => {
                            if requested_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requestedAt"));
                            }
                            requested_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Response => {
                            if response__.is_some() {
                                return Err(serde::de::Error::duplicate_field("response"));
                            }
                            response__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Challenge {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    requester: requester__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    benchmark_type: benchmark_type__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    requested_at: requested_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    response: response__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.Challenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ChallengeCompletedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.provider.is_empty() {
            len += 1;
        }
        if self.passed {
            len += 1;
        }
        if self.completed_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ChallengeCompletedEvent", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if self.passed {
            struct_ser.serialize_field("passed", &self.passed)?;
        }
        if self.completed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("completedAt", ToString::to_string(&self.completed_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ChallengeCompletedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "provider",
            "passed",
            "completed_at",
            "completedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            Provider,
            Passed,
            CompletedAt,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "provider" => Ok(GeneratedField::Provider),
                            "passed" => Ok(GeneratedField::Passed),
                            "completedAt" | "completed_at" => Ok(GeneratedField::CompletedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ChallengeCompletedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ChallengeCompletedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ChallengeCompletedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut provider__ = None;
                let mut passed__ = None;
                let mut completed_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Passed => {
                            if passed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("passed"));
                            }
                            passed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CompletedAt => {
                            if completed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completedAt"));
                            }
                            completed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ChallengeCompletedEvent {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    passed: passed__.unwrap_or_default(),
                    completed_at: completed_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ChallengeCompletedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ChallengeExpiredEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.provider.is_empty() {
            len += 1;
        }
        if self.expired_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ChallengeExpiredEvent", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if self.expired_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiredAt", ToString::to_string(&self.expired_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ChallengeExpiredEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "provider",
            "expired_at",
            "expiredAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            Provider,
            ExpiredAt,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "provider" => Ok(GeneratedField::Provider),
                            "expiredAt" | "expired_at" => Ok(GeneratedField::ExpiredAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ChallengeExpiredEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ChallengeExpiredEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ChallengeExpiredEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut provider__ = None;
                let mut expired_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiredAt => {
                            if expired_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiredAt"));
                            }
                            expired_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ChallengeExpiredEvent {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    expired_at: expired_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ChallengeExpiredEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ChallengeRequestedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.requester.is_empty() {
            len += 1;
        }
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.benchmark_type.is_empty() {
            len += 1;
        }
        if self.requested_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ChallengeRequestedEvent", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.requester.is_empty() {
            struct_ser.serialize_field("requester", &self.requester)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.benchmark_type.is_empty() {
            struct_ser.serialize_field("benchmarkType", &self.benchmark_type)?;
        }
        if self.requested_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("requestedAt", ToString::to_string(&self.requested_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ChallengeRequestedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
            "requester",
            "provider",
            "benchmark_type",
            "benchmarkType",
            "requested_at",
            "requestedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
            Requester,
            Provider,
            BenchmarkType,
            RequestedAt,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "requester" => Ok(GeneratedField::Requester),
                            "provider" => Ok(GeneratedField::Provider),
                            "benchmarkType" | "benchmark_type" => Ok(GeneratedField::BenchmarkType),
                            "requestedAt" | "requested_at" => Ok(GeneratedField::RequestedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ChallengeRequestedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ChallengeRequestedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ChallengeRequestedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                let mut requester__ = None;
                let mut provider__ = None;
                let mut benchmark_type__ = None;
                let mut requested_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Requester => {
                            if requester__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requester"));
                            }
                            requester__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BenchmarkType => {
                            if benchmark_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("benchmarkType"));
                            }
                            benchmark_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequestedAt => {
                            if requested_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requestedAt"));
                            }
                            requested_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ChallengeRequestedEvent {
                    challenge_id: challenge_id__.unwrap_or_default(),
                    requester: requester__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    benchmark_type: benchmark_type__.unwrap_or_default(),
                    requested_at: requested_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ChallengeRequestedEvent", FIELDS, GeneratedVisitor)
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
        if !self.provider_benchmarks.is_empty() {
            len += 1;
        }
        if !self.challenges.is_empty() {
            len += 1;
        }
        if !self.anomaly_flags.is_empty() {
            len += 1;
        }
        if self.challenge_sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.provider_benchmarks.is_empty() {
            struct_ser.serialize_field("providerBenchmarks", &self.provider_benchmarks)?;
        }
        if !self.challenges.is_empty() {
            struct_ser.serialize_field("challenges", &self.challenges)?;
        }
        if !self.anomaly_flags.is_empty() {
            struct_ser.serialize_field("anomalyFlags", &self.anomaly_flags)?;
        }
        if self.challenge_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challengeSequence", ToString::to_string(&self.challenge_sequence).as_str())?;
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
            "provider_benchmarks",
            "providerBenchmarks",
            "challenges",
            "anomaly_flags",
            "anomalyFlags",
            "challenge_sequence",
            "challengeSequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            ProviderBenchmarks,
            Challenges,
            AnomalyFlags,
            ChallengeSequence,
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
                            "providerBenchmarks" | "provider_benchmarks" => Ok(GeneratedField::ProviderBenchmarks),
                            "challenges" => Ok(GeneratedField::Challenges),
                            "anomalyFlags" | "anomaly_flags" => Ok(GeneratedField::AnomalyFlags),
                            "challengeSequence" | "challenge_sequence" => Ok(GeneratedField::ChallengeSequence),
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
                formatter.write_str("struct virtengine.benchmark.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut provider_benchmarks__ = None;
                let mut challenges__ = None;
                let mut anomaly_flags__ = None;
                let mut challenge_sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::ProviderBenchmarks => {
                            if provider_benchmarks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerBenchmarks"));
                            }
                            provider_benchmarks__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Challenges => {
                            if challenges__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challenges"));
                            }
                            challenges__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AnomalyFlags => {
                            if anomaly_flags__.is_some() {
                                return Err(serde::de::Error::duplicate_field("anomalyFlags"));
                            }
                            anomaly_flags__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChallengeSequence => {
                            if challenge_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeSequence"));
                            }
                            challenge_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    provider_benchmarks: provider_benchmarks__.unwrap_or_default(),
                    challenges: challenges__.unwrap_or_default(),
                    anomaly_flags: anomaly_flags__.unwrap_or_default(),
                    challenge_sequence: challenge_sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgFlagProvider {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reporter.is_empty() {
            len += 1;
        }
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgFlagProvider", len)?;
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", &self.reporter)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
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
impl<'de> serde::Deserialize<'de> for MsgFlagProvider {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reporter",
            "provider",
            "reason",
            "evidence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reporter,
            Provider,
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
                            "reporter" => Ok(GeneratedField::Reporter),
                            "provider" => Ok(GeneratedField::Provider),
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
            type Value = MsgFlagProvider;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgFlagProvider")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgFlagProvider, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reporter__ = None;
                let mut provider__ = None;
                let mut reason__ = None;
                let mut evidence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
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
                Ok(MsgFlagProvider {
                    reporter: reporter__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgFlagProvider", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgFlagProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgFlagProviderResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgFlagProviderResponse {
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
            type Value = MsgFlagProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgFlagProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgFlagProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgFlagProviderResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgFlagProviderResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRequestChallenge {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.requester.is_empty() {
            len += 1;
        }
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.benchmark_type.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgRequestChallenge", len)?;
        if !self.requester.is_empty() {
            struct_ser.serialize_field("requester", &self.requester)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.benchmark_type.is_empty() {
            struct_ser.serialize_field("benchmarkType", &self.benchmark_type)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRequestChallenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "requester",
            "provider",
            "benchmark_type",
            "benchmarkType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Requester,
            Provider,
            BenchmarkType,
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
                            "requester" => Ok(GeneratedField::Requester),
                            "provider" => Ok(GeneratedField::Provider),
                            "benchmarkType" | "benchmark_type" => Ok(GeneratedField::BenchmarkType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRequestChallenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgRequestChallenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRequestChallenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut requester__ = None;
                let mut provider__ = None;
                let mut benchmark_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Requester => {
                            if requester__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requester"));
                            }
                            requester__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BenchmarkType => {
                            if benchmark_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("benchmarkType"));
                            }
                            benchmark_type__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRequestChallenge {
                    requester: requester__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    benchmark_type: benchmark_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgRequestChallenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRequestChallengeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgRequestChallengeResponse", len)?;
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRequestChallengeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "challenge_id",
            "challengeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeId,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRequestChallengeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgRequestChallengeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRequestChallengeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRequestChallengeResponse {
                    challenge_id: challenge_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgRequestChallengeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveAnomalyFlag {
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
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.resolution.is_empty() {
            len += 1;
        }
        if self.is_valid {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgResolveAnomalyFlag", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.resolution.is_empty() {
            struct_ser.serialize_field("resolution", &self.resolution)?;
        }
        if self.is_valid {
            struct_ser.serialize_field("isValid", &self.is_valid)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveAnomalyFlag {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "provider",
            "resolution",
            "is_valid",
            "isValid",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            Provider,
            Resolution,
            IsValid,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "isValid" | "is_valid" => Ok(GeneratedField::IsValid),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgResolveAnomalyFlag;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgResolveAnomalyFlag")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveAnomalyFlag, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut provider__ = None;
                let mut resolution__ = None;
                let mut is_valid__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsValid => {
                            if is_valid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isValid"));
                            }
                            is_valid__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgResolveAnomalyFlag {
                    authority: authority__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    is_valid: is_valid__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgResolveAnomalyFlag", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveAnomalyFlagResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgResolveAnomalyFlagResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveAnomalyFlagResponse {
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
            type Value = MsgResolveAnomalyFlagResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgResolveAnomalyFlagResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveAnomalyFlagResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgResolveAnomalyFlagResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgResolveAnomalyFlagResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRespondChallenge {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if self.result.is_some() {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgRespondChallenge", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if let Some(v) = self.result.as_ref() {
            struct_ser.serialize_field("result", v)?;
        }
        if !self.signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signature", pbjson::private::base64::encode(&self.signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRespondChallenge {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "challenge_id",
            "challengeId",
            "result",
            "signature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            ChallengeId,
            Result,
            Signature,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "result" => Ok(GeneratedField::Result),
                            "signature" => Ok(GeneratedField::Signature),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRespondChallenge;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgRespondChallenge")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRespondChallenge, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut challenge_id__ = None;
                let mut result__ = None;
                let mut signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Result => {
                            if result__.is_some() {
                                return Err(serde::de::Error::duplicate_field("result"));
                            }
                            result__ = map_.next_value()?;
                        }
                        GeneratedField::Signature => {
                            if signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signature"));
                            }
                            signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRespondChallenge {
                    provider: provider__.unwrap_or_default(),
                    challenge_id: challenge_id__.unwrap_or_default(),
                    result: result__,
                    signature: signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgRespondChallenge", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRespondChallengeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgRespondChallengeResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRespondChallengeResponse {
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
            type Value = MsgRespondChallengeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgRespondChallengeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRespondChallengeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRespondChallengeResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgRespondChallengeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitBenchmarks {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.cluster_id.is_empty() {
            len += 1;
        }
        if !self.results.is_empty() {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgSubmitBenchmarks", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.cluster_id.is_empty() {
            struct_ser.serialize_field("clusterId", &self.cluster_id)?;
        }
        if !self.results.is_empty() {
            struct_ser.serialize_field("results", &self.results)?;
        }
        if !self.signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signature", pbjson::private::base64::encode(&self.signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitBenchmarks {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "cluster_id",
            "clusterId",
            "results",
            "signature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            ClusterId,
            Results,
            Signature,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "clusterId" | "cluster_id" => Ok(GeneratedField::ClusterId),
                            "results" => Ok(GeneratedField::Results),
                            "signature" => Ok(GeneratedField::Signature),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitBenchmarks;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgSubmitBenchmarks")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitBenchmarks, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut cluster_id__ = None;
                let mut results__ = None;
                let mut signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClusterId => {
                            if cluster_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clusterId"));
                            }
                            cluster_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Results => {
                            if results__.is_some() {
                                return Err(serde::de::Error::duplicate_field("results"));
                            }
                            results__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Signature => {
                            if signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signature"));
                            }
                            signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgSubmitBenchmarks {
                    provider: provider__.unwrap_or_default(),
                    cluster_id: cluster_id__.unwrap_or_default(),
                    results: results__.unwrap_or_default(),
                    signature: signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgSubmitBenchmarks", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitBenchmarksResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgSubmitBenchmarksResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitBenchmarksResponse {
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
            type Value = MsgSubmitBenchmarksResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgSubmitBenchmarksResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitBenchmarksResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgSubmitBenchmarksResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgSubmitBenchmarksResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUnflagProvider {
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
        if !self.provider.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgUnflagProvider", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUnflagProvider {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "provider",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            Provider,
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
                            "provider" => Ok(GeneratedField::Provider),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUnflagProvider;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgUnflagProvider")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUnflagProvider, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut provider__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUnflagProvider {
                    authority: authority__.unwrap_or_default(),
                    provider: provider__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgUnflagProvider", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUnflagProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.MsgUnflagProviderResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUnflagProviderResponse {
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
            type Value = MsgUnflagProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.MsgUnflagProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUnflagProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUnflagProviderResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.MsgUnflagProviderResponse", FIELDS, GeneratedVisitor)
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
        if self.challenge_timeout != 0 {
            len += 1;
        }
        if self.benchmark_validity_period != 0 {
            len += 1;
        }
        if !self.min_benchmark_score.is_empty() {
            len += 1;
        }
        if self.max_anomaly_flags != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.Params", len)?;
        if self.challenge_timeout != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("challengeTimeout", ToString::to_string(&self.challenge_timeout).as_str())?;
        }
        if self.benchmark_validity_period != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("benchmarkValidityPeriod", ToString::to_string(&self.benchmark_validity_period).as_str())?;
        }
        if !self.min_benchmark_score.is_empty() {
            struct_ser.serialize_field("minBenchmarkScore", &self.min_benchmark_score)?;
        }
        if self.max_anomaly_flags != 0 {
            struct_ser.serialize_field("maxAnomalyFlags", &self.max_anomaly_flags)?;
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
            "challenge_timeout",
            "challengeTimeout",
            "benchmark_validity_period",
            "benchmarkValidityPeriod",
            "min_benchmark_score",
            "minBenchmarkScore",
            "max_anomaly_flags",
            "maxAnomalyFlags",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChallengeTimeout,
            BenchmarkValidityPeriod,
            MinBenchmarkScore,
            MaxAnomalyFlags,
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
                            "challengeTimeout" | "challenge_timeout" => Ok(GeneratedField::ChallengeTimeout),
                            "benchmarkValidityPeriod" | "benchmark_validity_period" => Ok(GeneratedField::BenchmarkValidityPeriod),
                            "minBenchmarkScore" | "min_benchmark_score" => Ok(GeneratedField::MinBenchmarkScore),
                            "maxAnomalyFlags" | "max_anomaly_flags" => Ok(GeneratedField::MaxAnomalyFlags),
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
                formatter.write_str("struct virtengine.benchmark.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut challenge_timeout__ = None;
                let mut benchmark_validity_period__ = None;
                let mut min_benchmark_score__ = None;
                let mut max_anomaly_flags__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ChallengeTimeout => {
                            if challenge_timeout__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeTimeout"));
                            }
                            challenge_timeout__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BenchmarkValidityPeriod => {
                            if benchmark_validity_period__.is_some() {
                                return Err(serde::de::Error::duplicate_field("benchmarkValidityPeriod"));
                            }
                            benchmark_validity_period__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinBenchmarkScore => {
                            if min_benchmark_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minBenchmarkScore"));
                            }
                            min_benchmark_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxAnomalyFlags => {
                            if max_anomaly_flags__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxAnomalyFlags"));
                            }
                            max_anomaly_flags__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Params {
                    challenge_timeout: challenge_timeout__.unwrap_or_default(),
                    benchmark_validity_period: benchmark_validity_period__.unwrap_or_default(),
                    min_benchmark_score: min_benchmark_score__.unwrap_or_default(),
                    max_anomaly_flags: max_anomaly_flags__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProviderBenchmark {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.results.is_empty() {
            len += 1;
        }
        if !self.reliability_score.is_empty() {
            len += 1;
        }
        if self.last_updated != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ProviderBenchmark", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.results.is_empty() {
            struct_ser.serialize_field("results", &self.results)?;
        }
        if !self.reliability_score.is_empty() {
            struct_ser.serialize_field("reliabilityScore", &self.reliability_score)?;
        }
        if self.last_updated != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdated", ToString::to_string(&self.last_updated).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProviderBenchmark {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "results",
            "reliability_score",
            "reliabilityScore",
            "last_updated",
            "lastUpdated",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            Results,
            ReliabilityScore,
            LastUpdated,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "results" => Ok(GeneratedField::Results),
                            "reliabilityScore" | "reliability_score" => Ok(GeneratedField::ReliabilityScore),
                            "lastUpdated" | "last_updated" => Ok(GeneratedField::LastUpdated),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProviderBenchmark;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ProviderBenchmark")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ProviderBenchmark, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut results__ = None;
                let mut reliability_score__ = None;
                let mut last_updated__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Results => {
                            if results__.is_some() {
                                return Err(serde::de::Error::duplicate_field("results"));
                            }
                            results__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReliabilityScore => {
                            if reliability_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reliabilityScore"));
                            }
                            reliability_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastUpdated => {
                            if last_updated__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdated"));
                            }
                            last_updated__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ProviderBenchmark {
                    provider: provider__.unwrap_or_default(),
                    results: results__.unwrap_or_default(),
                    reliability_score: reliability_score__.unwrap_or_default(),
                    last_updated: last_updated__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ProviderBenchmark", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProviderFlaggedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.reporter.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if self.flagged_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ProviderFlaggedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", &self.reporter)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if self.flagged_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("flaggedAt", ToString::to_string(&self.flagged_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProviderFlaggedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "reporter",
            "reason",
            "flagged_at",
            "flaggedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            Reporter,
            Reason,
            FlaggedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "reporter" => Ok(GeneratedField::Reporter),
                            "reason" => Ok(GeneratedField::Reason),
                            "flaggedAt" | "flagged_at" => Ok(GeneratedField::FlaggedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProviderFlaggedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ProviderFlaggedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ProviderFlaggedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut reporter__ = None;
                let mut reason__ = None;
                let mut flagged_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FlaggedAt => {
                            if flagged_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("flaggedAt"));
                            }
                            flagged_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ProviderFlaggedEvent {
                    provider: provider__.unwrap_or_default(),
                    reporter: reporter__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    flagged_at: flagged_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ProviderFlaggedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProviderUnflaggedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if self.unflagged_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ProviderUnflaggedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if self.unflagged_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("unflaggedAt", ToString::to_string(&self.unflagged_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProviderUnflaggedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "unflagged_at",
            "unflaggedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            UnflaggedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "unflaggedAt" | "unflagged_at" => Ok(GeneratedField::UnflaggedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProviderUnflaggedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ProviderUnflaggedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ProviderUnflaggedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut unflagged_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UnflaggedAt => {
                            if unflagged_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unflaggedAt"));
                            }
                            unflagged_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ProviderUnflaggedEvent {
                    provider: provider__.unwrap_or_default(),
                    unflagged_at: unflagged_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ProviderUnflaggedEvent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ReliabilityScoreUpdatedEvent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.old_score.is_empty() {
            len += 1;
        }
        if !self.new_score.is_empty() {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.benchmark.v1.ReliabilityScoreUpdatedEvent", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.old_score.is_empty() {
            struct_ser.serialize_field("oldScore", &self.old_score)?;
        }
        if !self.new_score.is_empty() {
            struct_ser.serialize_field("newScore", &self.new_score)?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ReliabilityScoreUpdatedEvent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "old_score",
            "oldScore",
            "new_score",
            "newScore",
            "updated_at",
            "updatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            OldScore,
            NewScore,
            UpdatedAt,
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
                            "provider" => Ok(GeneratedField::Provider),
                            "oldScore" | "old_score" => Ok(GeneratedField::OldScore),
                            "newScore" | "new_score" => Ok(GeneratedField::NewScore),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ReliabilityScoreUpdatedEvent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.benchmark.v1.ReliabilityScoreUpdatedEvent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ReliabilityScoreUpdatedEvent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut old_score__ = None;
                let mut new_score__ = None;
                let mut updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OldScore => {
                            if old_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldScore"));
                            }
                            old_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewScore => {
                            if new_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newScore"));
                            }
                            new_score__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ReliabilityScoreUpdatedEvent {
                    provider: provider__.unwrap_or_default(),
                    old_score: old_score__.unwrap_or_default(),
                    new_score: new_score__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.benchmark.v1.ReliabilityScoreUpdatedEvent", FIELDS, GeneratedVisitor)
    }
}

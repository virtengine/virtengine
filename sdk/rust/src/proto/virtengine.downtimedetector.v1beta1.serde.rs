// @generated
impl serde::Serialize for Downtime {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Duration30s => "DURATION_30S",
            Self::Duration1m => "DURATION_1M",
            Self::Duration2m => "DURATION_2M",
            Self::Duration3m => "DURATION_3M",
            Self::Duration4m => "DURATION_4M",
            Self::Duration5m => "DURATION_5M",
            Self::Duration10m => "DURATION_10M",
            Self::Duration20m => "DURATION_20M",
            Self::Duration30m => "DURATION_30M",
            Self::Duration40m => "DURATION_40M",
            Self::Duration50m => "DURATION_50M",
            Self::Duration1h => "DURATION_1H",
            Self::Duration15h => "DURATION_1_5H",
            Self::Duration2h => "DURATION_2H",
            Self::Duration25h => "DURATION_2_5H",
            Self::Duration3h => "DURATION_3H",
            Self::Duration4h => "DURATION_4H",
            Self::Duration5h => "DURATION_5H",
            Self::Duration6h => "DURATION_6H",
            Self::Duration9h => "DURATION_9H",
            Self::Duration12h => "DURATION_12H",
            Self::Duration18h => "DURATION_18H",
            Self::Duration24h => "DURATION_24H",
            Self::Duration36h => "DURATION_36H",
            Self::Duration48h => "DURATION_48H",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for Downtime {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "DURATION_30S",
            "DURATION_1M",
            "DURATION_2M",
            "DURATION_3M",
            "DURATION_4M",
            "DURATION_5M",
            "DURATION_10M",
            "DURATION_20M",
            "DURATION_30M",
            "DURATION_40M",
            "DURATION_50M",
            "DURATION_1H",
            "DURATION_1_5H",
            "DURATION_2H",
            "DURATION_2_5H",
            "DURATION_3H",
            "DURATION_4H",
            "DURATION_5H",
            "DURATION_6H",
            "DURATION_9H",
            "DURATION_12H",
            "DURATION_18H",
            "DURATION_24H",
            "DURATION_36H",
            "DURATION_48H",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Downtime;

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
                    "DURATION_30S" => Ok(Downtime::Duration30s),
                    "DURATION_1M" => Ok(Downtime::Duration1m),
                    "DURATION_2M" => Ok(Downtime::Duration2m),
                    "DURATION_3M" => Ok(Downtime::Duration3m),
                    "DURATION_4M" => Ok(Downtime::Duration4m),
                    "DURATION_5M" => Ok(Downtime::Duration5m),
                    "DURATION_10M" => Ok(Downtime::Duration10m),
                    "DURATION_20M" => Ok(Downtime::Duration20m),
                    "DURATION_30M" => Ok(Downtime::Duration30m),
                    "DURATION_40M" => Ok(Downtime::Duration40m),
                    "DURATION_50M" => Ok(Downtime::Duration50m),
                    "DURATION_1H" => Ok(Downtime::Duration1h),
                    "DURATION_1_5H" => Ok(Downtime::Duration15h),
                    "DURATION_2H" => Ok(Downtime::Duration2h),
                    "DURATION_2_5H" => Ok(Downtime::Duration25h),
                    "DURATION_3H" => Ok(Downtime::Duration3h),
                    "DURATION_4H" => Ok(Downtime::Duration4h),
                    "DURATION_5H" => Ok(Downtime::Duration5h),
                    "DURATION_6H" => Ok(Downtime::Duration6h),
                    "DURATION_9H" => Ok(Downtime::Duration9h),
                    "DURATION_12H" => Ok(Downtime::Duration12h),
                    "DURATION_18H" => Ok(Downtime::Duration18h),
                    "DURATION_24H" => Ok(Downtime::Duration24h),
                    "DURATION_36H" => Ok(Downtime::Duration36h),
                    "DURATION_48H" => Ok(Downtime::Duration48h),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for GenesisDowntimeEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.duration != 0 {
            len += 1;
        }
        if self.last_downtime.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.downtimedetector.v1beta1.GenesisDowntimeEntry", len)?;
        if self.duration != 0 {
            let v = Downtime::try_from(self.duration)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.duration)))?;
            struct_ser.serialize_field("duration", &v)?;
        }
        if let Some(v) = self.last_downtime.as_ref() {
            struct_ser.serialize_field("lastDowntime", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GenesisDowntimeEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "duration",
            "last_downtime",
            "lastDowntime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Duration,
            LastDowntime,
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
                            "duration" => Ok(GeneratedField::Duration),
                            "lastDowntime" | "last_downtime" => Ok(GeneratedField::LastDowntime),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GenesisDowntimeEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.downtimedetector.v1beta1.GenesisDowntimeEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisDowntimeEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut duration__ = None;
                let mut last_downtime__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Duration => {
                            if duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("duration"));
                            }
                            duration__ = Some(map_.next_value::<Downtime>()? as i32);
                        }
                        GeneratedField::LastDowntime => {
                            if last_downtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastDowntime"));
                            }
                            last_downtime__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisDowntimeEntry {
                    duration: duration__.unwrap_or_default(),
                    last_downtime: last_downtime__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.downtimedetector.v1beta1.GenesisDowntimeEntry", FIELDS, GeneratedVisitor)
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
        if !self.downtimes.is_empty() {
            len += 1;
        }
        if self.last_block_time.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.downtimedetector.v1beta1.GenesisState", len)?;
        if !self.downtimes.is_empty() {
            struct_ser.serialize_field("downtimes", &self.downtimes)?;
        }
        if let Some(v) = self.last_block_time.as_ref() {
            struct_ser.serialize_field("lastBlockTime", v)?;
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
            "downtimes",
            "last_block_time",
            "lastBlockTime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Downtimes,
            LastBlockTime,
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
                            "downtimes" => Ok(GeneratedField::Downtimes),
                            "lastBlockTime" | "last_block_time" => Ok(GeneratedField::LastBlockTime),
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
                formatter.write_str("struct virtengine.downtimedetector.v1beta1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut downtimes__ = None;
                let mut last_block_time__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Downtimes => {
                            if downtimes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("downtimes"));
                            }
                            downtimes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastBlockTime => {
                            if last_block_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastBlockTime"));
                            }
                            last_block_time__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisState {
                    downtimes: downtimes__.unwrap_or_default(),
                    last_block_time: last_block_time__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.downtimedetector.v1beta1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RecoveredSinceDowntimeOfLengthRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.downtime != 0 {
            len += 1;
        }
        if self.recovery.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest", len)?;
        if self.downtime != 0 {
            let v = Downtime::try_from(self.downtime)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.downtime)))?;
            struct_ser.serialize_field("downtime", &v)?;
        }
        if let Some(v) = self.recovery.as_ref() {
            struct_ser.serialize_field("recovery", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RecoveredSinceDowntimeOfLengthRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "downtime",
            "recovery",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Downtime,
            Recovery,
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
                            "downtime" => Ok(GeneratedField::Downtime),
                            "recovery" => Ok(GeneratedField::Recovery),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RecoveredSinceDowntimeOfLengthRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RecoveredSinceDowntimeOfLengthRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut downtime__ = None;
                let mut recovery__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Downtime => {
                            if downtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("downtime"));
                            }
                            downtime__ = Some(map_.next_value::<Downtime>()? as i32);
                        }
                        GeneratedField::Recovery => {
                            if recovery__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recovery"));
                            }
                            recovery__ = map_.next_value()?;
                        }
                    }
                }
                Ok(RecoveredSinceDowntimeOfLengthRequest {
                    downtime: downtime__.unwrap_or_default(),
                    recovery: recovery__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RecoveredSinceDowntimeOfLengthResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.succesfully_recovered {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse", len)?;
        if self.succesfully_recovered {
            struct_ser.serialize_field("succesfullyRecovered", &self.succesfully_recovered)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RecoveredSinceDowntimeOfLengthResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "succesfully_recovered",
            "succesfullyRecovered",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SuccesfullyRecovered,
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
                            "succesfullyRecovered" | "succesfully_recovered" => Ok(GeneratedField::SuccesfullyRecovered),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RecoveredSinceDowntimeOfLengthResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RecoveredSinceDowntimeOfLengthResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut succesfully_recovered__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SuccesfullyRecovered => {
                            if succesfully_recovered__.is_some() {
                                return Err(serde::de::Error::duplicate_field("succesfullyRecovered"));
                            }
                            succesfully_recovered__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(RecoveredSinceDowntimeOfLengthResponse {
                    succesfully_recovered: succesfully_recovered__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.downtimedetector.v1beta1.RecoveredSinceDowntimeOfLengthResponse", FIELDS, GeneratedVisitor)
    }
}

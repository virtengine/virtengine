// @generated
impl serde::Serialize for EpochInfo {
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
        if self.start_time.is_some() {
            len += 1;
        }
        if self.duration.is_some() {
            len += 1;
        }
        if self.current_epoch != 0 {
            len += 1;
        }
        if self.current_epoch_start_time.is_some() {
            len += 1;
        }
        if self.epoch_counting_started {
            len += 1;
        }
        if self.current_epoch_start_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.EpochInfo", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if let Some(v) = self.start_time.as_ref() {
            struct_ser.serialize_field("startTime", v)?;
        }
        if let Some(v) = self.duration.as_ref() {
            struct_ser.serialize_field("duration", v)?;
        }
        if self.current_epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("currentEpoch", ToString::to_string(&self.current_epoch).as_str())?;
        }
        if let Some(v) = self.current_epoch_start_time.as_ref() {
            struct_ser.serialize_field("currentEpochStartTime", v)?;
        }
        if self.epoch_counting_started {
            struct_ser.serialize_field("epochCountingStarted", &self.epoch_counting_started)?;
        }
        if self.current_epoch_start_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("currentEpochStartHeight", ToString::to_string(&self.current_epoch_start_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EpochInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "start_time",
            "startTime",
            "duration",
            "current_epoch",
            "currentEpoch",
            "current_epoch_start_time",
            "currentEpochStartTime",
            "epoch_counting_started",
            "epochCountingStarted",
            "current_epoch_start_height",
            "currentEpochStartHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            StartTime,
            Duration,
            CurrentEpoch,
            CurrentEpochStartTime,
            EpochCountingStarted,
            CurrentEpochStartHeight,
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
                            "startTime" | "start_time" => Ok(GeneratedField::StartTime),
                            "duration" => Ok(GeneratedField::Duration),
                            "currentEpoch" | "current_epoch" => Ok(GeneratedField::CurrentEpoch),
                            "currentEpochStartTime" | "current_epoch_start_time" => Ok(GeneratedField::CurrentEpochStartTime),
                            "epochCountingStarted" | "epoch_counting_started" => Ok(GeneratedField::EpochCountingStarted),
                            "currentEpochStartHeight" | "current_epoch_start_height" => Ok(GeneratedField::CurrentEpochStartHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EpochInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.EpochInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EpochInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut start_time__ = None;
                let mut duration__ = None;
                let mut current_epoch__ = None;
                let mut current_epoch_start_time__ = None;
                let mut epoch_counting_started__ = None;
                let mut current_epoch_start_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StartTime => {
                            if start_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("startTime"));
                            }
                            start_time__ = map_.next_value()?;
                        }
                        GeneratedField::Duration => {
                            if duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("duration"));
                            }
                            duration__ = map_.next_value()?;
                        }
                        GeneratedField::CurrentEpoch => {
                            if current_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentEpoch"));
                            }
                            current_epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CurrentEpochStartTime => {
                            if current_epoch_start_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentEpochStartTime"));
                            }
                            current_epoch_start_time__ = map_.next_value()?;
                        }
                        GeneratedField::EpochCountingStarted => {
                            if epoch_counting_started__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochCountingStarted"));
                            }
                            epoch_counting_started__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CurrentEpochStartHeight => {
                            if current_epoch_start_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentEpochStartHeight"));
                            }
                            current_epoch_start_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EpochInfo {
                    id: id__.unwrap_or_default(),
                    start_time: start_time__,
                    duration: duration__,
                    current_epoch: current_epoch__.unwrap_or_default(),
                    current_epoch_start_time: current_epoch_start_time__,
                    epoch_counting_started: epoch_counting_started__.unwrap_or_default(),
                    current_epoch_start_height: current_epoch_start_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.EpochInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEpochEnd {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.epoch_number != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.EventEpochEnd", len)?;
        if self.epoch_number != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEpochEnd {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_number",
            "epochNumber",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNumber,
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
                            "epochNumber" | "epoch_number" => Ok(GeneratedField::EpochNumber),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventEpochEnd;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.EventEpochEnd")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEpochEnd, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_number__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EpochNumber => {
                            if epoch_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNumber"));
                            }
                            epoch_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventEpochEnd {
                    epoch_number: epoch_number__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.EventEpochEnd", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventEpochStart {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.epoch_number != 0 {
            len += 1;
        }
        if self.epoch_start_time != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.EventEpochStart", len)?;
        if self.epoch_number != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if self.epoch_start_time != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochStartTime", ToString::to_string(&self.epoch_start_time).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventEpochStart {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_number",
            "epochNumber",
            "epoch_start_time",
            "epochStartTime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNumber,
            EpochStartTime,
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
                            "epochNumber" | "epoch_number" => Ok(GeneratedField::EpochNumber),
                            "epochStartTime" | "epoch_start_time" => Ok(GeneratedField::EpochStartTime),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventEpochStart;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.EventEpochStart")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventEpochStart, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_number__ = None;
                let mut epoch_start_time__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EpochNumber => {
                            if epoch_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNumber"));
                            }
                            epoch_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EpochStartTime => {
                            if epoch_start_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochStartTime"));
                            }
                            epoch_start_time__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventEpochStart {
                    epoch_number: epoch_number__.unwrap_or_default(),
                    epoch_start_time: epoch_start_time__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.EventEpochStart", FIELDS, GeneratedVisitor)
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
        if !self.epochs.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.GenesisState", len)?;
        if !self.epochs.is_empty() {
            struct_ser.serialize_field("epochs", &self.epochs)?;
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
            "epochs",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Epochs,
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
                            "epochs" => Ok(GeneratedField::Epochs),
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
                formatter.write_str("struct virtengine.epochs.v1beta1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epochs__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Epochs => {
                            if epochs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochs"));
                            }
                            epochs__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    epochs: epochs__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryCurrentEpochRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.identifier.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.QueryCurrentEpochRequest", len)?;
        if !self.identifier.is_empty() {
            struct_ser.serialize_field("identifier", &self.identifier)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryCurrentEpochRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "identifier",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Identifier,
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
                            "identifier" => Ok(GeneratedField::Identifier),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryCurrentEpochRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.QueryCurrentEpochRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryCurrentEpochRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut identifier__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Identifier => {
                            if identifier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identifier"));
                            }
                            identifier__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryCurrentEpochRequest {
                    identifier: identifier__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.QueryCurrentEpochRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryCurrentEpochResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.current_epoch != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.QueryCurrentEpochResponse", len)?;
        if self.current_epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("currentEpoch", ToString::to_string(&self.current_epoch).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryCurrentEpochResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "current_epoch",
            "currentEpoch",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CurrentEpoch,
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
                            "currentEpoch" | "current_epoch" => Ok(GeneratedField::CurrentEpoch),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryCurrentEpochResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.QueryCurrentEpochResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryCurrentEpochResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut current_epoch__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CurrentEpoch => {
                            if current_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentEpoch"));
                            }
                            current_epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryCurrentEpochResponse {
                    current_epoch: current_epoch__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.QueryCurrentEpochResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryEpochInfosRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.QueryEpochInfosRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryEpochInfosRequest {
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
            type Value = QueryEpochInfosRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.QueryEpochInfosRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryEpochInfosRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryEpochInfosRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.QueryEpochInfosRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryEpochInfosResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.epochs.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.epochs.v1beta1.QueryEpochInfosResponse", len)?;
        if !self.epochs.is_empty() {
            struct_ser.serialize_field("epochs", &self.epochs)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryEpochInfosResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epochs",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Epochs,
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
                            "epochs" => Ok(GeneratedField::Epochs),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryEpochInfosResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.epochs.v1beta1.QueryEpochInfosResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryEpochInfosResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epochs__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Epochs => {
                            if epochs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochs"));
                            }
                            epochs__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryEpochInfosResponse {
                    epochs: epochs__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.epochs.v1beta1.QueryEpochInfosResponse", FIELDS, GeneratedVisitor)
    }
}

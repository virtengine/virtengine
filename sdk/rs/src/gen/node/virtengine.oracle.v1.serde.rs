// @generated
impl serde::Serialize for AggregatedPrice {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.denom.is_empty() {
            len += 1;
        }
        if !self.twap.is_empty() {
            len += 1;
        }
        if !self.median_price.is_empty() {
            len += 1;
        }
        if !self.min_price.is_empty() {
            len += 1;
        }
        if !self.max_price.is_empty() {
            len += 1;
        }
        if self.timestamp.is_some() {
            len += 1;
        }
        if self.num_sources != 0 {
            len += 1;
        }
        if self.deviation_bps != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.AggregatedPrice", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if !self.twap.is_empty() {
            struct_ser.serialize_field("twap", &self.twap)?;
        }
        if !self.median_price.is_empty() {
            struct_ser.serialize_field("medianPrice", &self.median_price)?;
        }
        if !self.min_price.is_empty() {
            struct_ser.serialize_field("minPrice", &self.min_price)?;
        }
        if !self.max_price.is_empty() {
            struct_ser.serialize_field("maxPrice", &self.max_price)?;
        }
        if let Some(v) = self.timestamp.as_ref() {
            struct_ser.serialize_field("timestamp", v)?;
        }
        if self.num_sources != 0 {
            struct_ser.serialize_field("numSources", &self.num_sources)?;
        }
        if self.deviation_bps != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("deviationBps", ToString::to_string(&self.deviation_bps).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AggregatedPrice {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
            "twap",
            "median_price",
            "medianPrice",
            "min_price",
            "minPrice",
            "max_price",
            "maxPrice",
            "timestamp",
            "num_sources",
            "numSources",
            "deviation_bps",
            "deviationBps",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
            Twap,
            MedianPrice,
            MinPrice,
            MaxPrice,
            Timestamp,
            NumSources,
            DeviationBps,
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
                            "denom" => Ok(GeneratedField::Denom),
                            "twap" => Ok(GeneratedField::Twap),
                            "medianPrice" | "median_price" => Ok(GeneratedField::MedianPrice),
                            "minPrice" | "min_price" => Ok(GeneratedField::MinPrice),
                            "maxPrice" | "max_price" => Ok(GeneratedField::MaxPrice),
                            "timestamp" => Ok(GeneratedField::Timestamp),
                            "numSources" | "num_sources" => Ok(GeneratedField::NumSources),
                            "deviationBps" | "deviation_bps" => Ok(GeneratedField::DeviationBps),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AggregatedPrice;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.AggregatedPrice")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AggregatedPrice, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                let mut twap__ = None;
                let mut median_price__ = None;
                let mut min_price__ = None;
                let mut max_price__ = None;
                let mut timestamp__ = None;
                let mut num_sources__ = None;
                let mut deviation_bps__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Twap => {
                            if twap__.is_some() {
                                return Err(serde::de::Error::duplicate_field("twap"));
                            }
                            twap__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MedianPrice => {
                            if median_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("medianPrice"));
                            }
                            median_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinPrice => {
                            if min_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minPrice"));
                            }
                            min_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MaxPrice => {
                            if max_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxPrice"));
                            }
                            max_price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = map_.next_value()?;
                        }
                        GeneratedField::NumSources => {
                            if num_sources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("numSources"));
                            }
                            num_sources__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DeviationBps => {
                            if deviation_bps__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviationBps"));
                            }
                            deviation_bps__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AggregatedPrice {
                    denom: denom__.unwrap_or_default(),
                    twap: twap__.unwrap_or_default(),
                    median_price: median_price__.unwrap_or_default(),
                    min_price: min_price__.unwrap_or_default(),
                    max_price: max_price__.unwrap_or_default(),
                    timestamp: timestamp__,
                    num_sources: num_sources__.unwrap_or_default(),
                    deviation_bps: deviation_bps__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.AggregatedPrice", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for DataId {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.denom.is_empty() {
            len += 1;
        }
        if !self.base_denom.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.DataID", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if !self.base_denom.is_empty() {
            struct_ser.serialize_field("baseDenom", &self.base_denom)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DataId {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
            "base_denom",
            "baseDenom",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
            BaseDenom,
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
                            "denom" => Ok(GeneratedField::Denom),
                            "baseDenom" | "base_denom" => Ok(GeneratedField::BaseDenom),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DataId;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.DataID")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DataId, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                let mut base_denom__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BaseDenom => {
                            if base_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("baseDenom"));
                            }
                            base_denom__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(DataId {
                    denom: denom__.unwrap_or_default(),
                    base_denom: base_denom__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.DataID", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventPriceData {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.source.is_empty() {
            len += 1;
        }
        if self.id.is_some() {
            len += 1;
        }
        if self.data.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.EventPriceData", len)?;
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.data.as_ref() {
            struct_ser.serialize_field("data", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventPriceData {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "source",
            "id",
            "data",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Source,
            Id,
            Data,
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
                            "source" => Ok(GeneratedField::Source),
                            "id" => Ok(GeneratedField::Id),
                            "data" => Ok(GeneratedField::Data),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventPriceData;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.EventPriceData")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventPriceData, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut source__ = None;
                let mut id__ = None;
                let mut data__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Data => {
                            if data__.is_some() {
                                return Err(serde::de::Error::duplicate_field("data"));
                            }
                            data__ = map_.next_value()?;
                        }
                    }
                }
                Ok(EventPriceData {
                    source: source__.unwrap_or_default(),
                    id: id__,
                    data: data__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.EventPriceData", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventPriceRecovered {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.source.is_empty() {
            len += 1;
        }
        if self.id.is_some() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.EventPriceRecovered", len)?;
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventPriceRecovered {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "source",
            "id",
            "height",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Source,
            Id,
            Height,
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
                            "source" => Ok(GeneratedField::Source),
                            "id" => Ok(GeneratedField::Id),
                            "height" => Ok(GeneratedField::Height),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventPriceRecovered;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.EventPriceRecovered")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventPriceRecovered, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut source__ = None;
                let mut id__ = None;
                let mut height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventPriceRecovered {
                    source: source__.unwrap_or_default(),
                    id: id__,
                    height: height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.EventPriceRecovered", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventPriceStaleWarning {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.source.is_empty() {
            len += 1;
        }
        if self.id.is_some() {
            len += 1;
        }
        if self.last_height != 0 {
            len += 1;
        }
        if self.blocks_to_stall != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.EventPriceStaleWarning", len)?;
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.last_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastHeight", ToString::to_string(&self.last_height).as_str())?;
        }
        if self.blocks_to_stall != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blocksToStall", ToString::to_string(&self.blocks_to_stall).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventPriceStaleWarning {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "source",
            "id",
            "last_height",
            "lastHeight",
            "blocks_to_stall",
            "blocksToStall",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Source,
            Id,
            LastHeight,
            BlocksToStall,
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
                            "source" => Ok(GeneratedField::Source),
                            "id" => Ok(GeneratedField::Id),
                            "lastHeight" | "last_height" => Ok(GeneratedField::LastHeight),
                            "blocksToStall" | "blocks_to_stall" => Ok(GeneratedField::BlocksToStall),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventPriceStaleWarning;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.EventPriceStaleWarning")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventPriceStaleWarning, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut source__ = None;
                let mut id__ = None;
                let mut last_height__ = None;
                let mut blocks_to_stall__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::LastHeight => {
                            if last_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastHeight"));
                            }
                            last_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlocksToStall => {
                            if blocks_to_stall__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blocksToStall"));
                            }
                            blocks_to_stall__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventPriceStaleWarning {
                    source: source__.unwrap_or_default(),
                    id: id__,
                    last_height: last_height__.unwrap_or_default(),
                    blocks_to_stall: blocks_to_stall__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.EventPriceStaleWarning", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventPriceStaled {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.source.is_empty() {
            len += 1;
        }
        if self.id.is_some() {
            len += 1;
        }
        if self.last_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.EventPriceStaled", len)?;
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.last_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastHeight", ToString::to_string(&self.last_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventPriceStaled {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "source",
            "id",
            "last_height",
            "lastHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Source,
            Id,
            LastHeight,
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
                            "source" => Ok(GeneratedField::Source),
                            "id" => Ok(GeneratedField::Id),
                            "lastHeight" | "last_height" => Ok(GeneratedField::LastHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventPriceStaled;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.EventPriceStaled")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventPriceStaled, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut source__ = None;
                let mut id__ = None;
                let mut last_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::LastHeight => {
                            if last_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastHeight"));
                            }
                            last_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(EventPriceStaled {
                    source: source__.unwrap_or_default(),
                    id: id__,
                    last_height: last_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.EventPriceStaled", FIELDS, GeneratedVisitor)
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
        if !self.prices.is_empty() {
            len += 1;
        }
        if !self.latest_height.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.prices.is_empty() {
            struct_ser.serialize_field("prices", &self.prices)?;
        }
        if !self.latest_height.is_empty() {
            struct_ser.serialize_field("latestHeight", &self.latest_height)?;
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
            "prices",
            "latest_height",
            "latestHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            Prices,
            LatestHeight,
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
                            "prices" => Ok(GeneratedField::Prices),
                            "latestHeight" | "latest_height" => Ok(GeneratedField::LatestHeight),
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
                formatter.write_str("struct virtengine.oracle.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut prices__ = None;
                let mut latest_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Prices => {
                            if prices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("prices"));
                            }
                            prices__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LatestHeight => {
                            if latest_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latestHeight"));
                            }
                            latest_height__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    prices: prices__.unwrap_or_default(),
                    latest_height: latest_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAddPriceEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.signer.is_empty() {
            len += 1;
        }
        if self.id.is_some() {
            len += 1;
        }
        if self.price.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.MsgAddPriceEntry", len)?;
        if !self.signer.is_empty() {
            struct_ser.serialize_field("signer", &self.signer)?;
        }
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.price.as_ref() {
            struct_ser.serialize_field("price", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAddPriceEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "signer",
            "id",
            "price",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Signer,
            Id,
            Price,
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
                            "signer" => Ok(GeneratedField::Signer),
                            "id" => Ok(GeneratedField::Id),
                            "price" => Ok(GeneratedField::Price),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAddPriceEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.MsgAddPriceEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAddPriceEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut signer__ = None;
                let mut id__ = None;
                let mut price__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Signer => {
                            if signer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signer"));
                            }
                            signer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Price => {
                            if price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("price"));
                            }
                            price__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgAddPriceEntry {
                    signer: signer__.unwrap_or_default(),
                    id: id__,
                    price: price__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.MsgAddPriceEntry", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAddPriceEntryResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.oracle.v1.MsgAddPriceEntryResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAddPriceEntryResponse {
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
            type Value = MsgAddPriceEntryResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.MsgAddPriceEntryResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAddPriceEntryResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgAddPriceEntryResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.MsgAddPriceEntryResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.oracle.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.oracle.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.oracle.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.oracle.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.oracle.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if !self.sources.is_empty() {
            len += 1;
        }
        if self.min_price_sources != 0 {
            len += 1;
        }
        if self.max_price_staleness_blocks != 0 {
            len += 1;
        }
        if self.twap_window != 0 {
            len += 1;
        }
        if self.max_price_deviation_bps != 0 {
            len += 1;
        }
        if !self.feed_contracts_params.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.Params", len)?;
        if !self.sources.is_empty() {
            struct_ser.serialize_field("sources", &self.sources)?;
        }
        if self.min_price_sources != 0 {
            struct_ser.serialize_field("minPriceSources", &self.min_price_sources)?;
        }
        if self.max_price_staleness_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxPriceStalenessBlocks", ToString::to_string(&self.max_price_staleness_blocks).as_str())?;
        }
        if self.twap_window != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("twapWindow", ToString::to_string(&self.twap_window).as_str())?;
        }
        if self.max_price_deviation_bps != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxPriceDeviationBps", ToString::to_string(&self.max_price_deviation_bps).as_str())?;
        }
        if !self.feed_contracts_params.is_empty() {
            struct_ser.serialize_field("feedContractsParams", &self.feed_contracts_params)?;
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
            "sources",
            "min_price_sources",
            "minPriceSources",
            "max_price_staleness_blocks",
            "maxPriceStalenessBlocks",
            "twap_window",
            "twapWindow",
            "max_price_deviation_bps",
            "maxPriceDeviationBps",
            "feed_contracts_params",
            "feedContractsParams",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sources,
            MinPriceSources,
            MaxPriceStalenessBlocks,
            TwapWindow,
            MaxPriceDeviationBps,
            FeedContractsParams,
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
                            "sources" => Ok(GeneratedField::Sources),
                            "minPriceSources" | "min_price_sources" => Ok(GeneratedField::MinPriceSources),
                            "maxPriceStalenessBlocks" | "max_price_staleness_blocks" => Ok(GeneratedField::MaxPriceStalenessBlocks),
                            "twapWindow" | "twap_window" => Ok(GeneratedField::TwapWindow),
                            "maxPriceDeviationBps" | "max_price_deviation_bps" => Ok(GeneratedField::MaxPriceDeviationBps),
                            "feedContractsParams" | "feed_contracts_params" => Ok(GeneratedField::FeedContractsParams),
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
                formatter.write_str("struct virtengine.oracle.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sources__ = None;
                let mut min_price_sources__ = None;
                let mut max_price_staleness_blocks__ = None;
                let mut twap_window__ = None;
                let mut max_price_deviation_bps__ = None;
                let mut feed_contracts_params__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sources => {
                            if sources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sources"));
                            }
                            sources__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinPriceSources => {
                            if min_price_sources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minPriceSources"));
                            }
                            min_price_sources__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxPriceStalenessBlocks => {
                            if max_price_staleness_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxPriceStalenessBlocks"));
                            }
                            max_price_staleness_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TwapWindow => {
                            if twap_window__.is_some() {
                                return Err(serde::de::Error::duplicate_field("twapWindow"));
                            }
                            twap_window__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxPriceDeviationBps => {
                            if max_price_deviation_bps__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxPriceDeviationBps"));
                            }
                            max_price_deviation_bps__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::FeedContractsParams => {
                            if feed_contracts_params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("feedContractsParams"));
                            }
                            feed_contracts_params__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Params {
                    sources: sources__.unwrap_or_default(),
                    min_price_sources: min_price_sources__.unwrap_or_default(),
                    max_price_staleness_blocks: max_price_staleness_blocks__.unwrap_or_default(),
                    twap_window: twap_window__.unwrap_or_default(),
                    max_price_deviation_bps: max_price_deviation_bps__.unwrap_or_default(),
                    feed_contracts_params: feed_contracts_params__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PriceData {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.id.is_some() {
            len += 1;
        }
        if self.state.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PriceData", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.state.as_ref() {
            struct_ser.serialize_field("state", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PriceData {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "state",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
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
                            "id" => Ok(GeneratedField::Id),
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
            type Value = PriceData;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PriceData")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PriceData, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut state__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = map_.next_value()?;
                        }
                    }
                }
                Ok(PriceData {
                    id: id__,
                    state: state__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PriceData", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PriceDataId {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.source != 0 {
            len += 1;
        }
        if !self.denom.is_empty() {
            len += 1;
        }
        if !self.base_denom.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PriceDataID", len)?;
        if self.source != 0 {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if !self.base_denom.is_empty() {
            struct_ser.serialize_field("baseDenom", &self.base_denom)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PriceDataId {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "source",
            "denom",
            "base_denom",
            "baseDenom",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Source,
            Denom,
            BaseDenom,
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
                            "source" => Ok(GeneratedField::Source),
                            "denom" => Ok(GeneratedField::Denom),
                            "baseDenom" | "base_denom" => Ok(GeneratedField::BaseDenom),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PriceDataId;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PriceDataID")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PriceDataId, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut source__ = None;
                let mut denom__ = None;
                let mut base_denom__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BaseDenom => {
                            if base_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("baseDenom"));
                            }
                            base_denom__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(PriceDataId {
                    source: source__.unwrap_or_default(),
                    denom: denom__.unwrap_or_default(),
                    base_denom: base_denom__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PriceDataID", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PriceDataRecordId {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.source != 0 {
            len += 1;
        }
        if !self.denom.is_empty() {
            len += 1;
        }
        if !self.base_denom.is_empty() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PriceDataRecordID", len)?;
        if self.source != 0 {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if !self.base_denom.is_empty() {
            struct_ser.serialize_field("baseDenom", &self.base_denom)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PriceDataRecordId {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "source",
            "denom",
            "base_denom",
            "baseDenom",
            "height",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Source,
            Denom,
            BaseDenom,
            Height,
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
                            "source" => Ok(GeneratedField::Source),
                            "denom" => Ok(GeneratedField::Denom),
                            "baseDenom" | "base_denom" => Ok(GeneratedField::BaseDenom),
                            "height" => Ok(GeneratedField::Height),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PriceDataRecordId;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PriceDataRecordID")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PriceDataRecordId, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut source__ = None;
                let mut denom__ = None;
                let mut base_denom__ = None;
                let mut height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BaseDenom => {
                            if base_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("baseDenom"));
                            }
                            base_denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(PriceDataRecordId {
                    source: source__.unwrap_or_default(),
                    denom: denom__.unwrap_or_default(),
                    base_denom: base_denom__.unwrap_or_default(),
                    height: height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PriceDataRecordID", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PriceDataState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.price.is_empty() {
            len += 1;
        }
        if self.timestamp.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PriceDataState", len)?;
        if !self.price.is_empty() {
            struct_ser.serialize_field("price", &self.price)?;
        }
        if let Some(v) = self.timestamp.as_ref() {
            struct_ser.serialize_field("timestamp", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PriceDataState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "price",
            "timestamp",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Price,
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
                            "price" => Ok(GeneratedField::Price),
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
            type Value = PriceDataState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PriceDataState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PriceDataState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut price__ = None;
                let mut timestamp__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Price => {
                            if price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("price"));
                            }
                            price__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = map_.next_value()?;
                        }
                    }
                }
                Ok(PriceDataState {
                    price: price__.unwrap_or_default(),
                    timestamp: timestamp__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PriceDataState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PriceHealth {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.denom.is_empty() {
            len += 1;
        }
        if self.is_healthy {
            len += 1;
        }
        if self.has_min_sources {
            len += 1;
        }
        if self.deviation_ok {
            len += 1;
        }
        if self.total_sources != 0 {
            len += 1;
        }
        if self.total_healthy_sources != 0 {
            len += 1;
        }
        if !self.failure_reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PriceHealth", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if self.is_healthy {
            struct_ser.serialize_field("isHealthy", &self.is_healthy)?;
        }
        if self.has_min_sources {
            struct_ser.serialize_field("hasMinSources", &self.has_min_sources)?;
        }
        if self.deviation_ok {
            struct_ser.serialize_field("deviationOk", &self.deviation_ok)?;
        }
        if self.total_sources != 0 {
            struct_ser.serialize_field("totalSources", &self.total_sources)?;
        }
        if self.total_healthy_sources != 0 {
            struct_ser.serialize_field("totalHealthySources", &self.total_healthy_sources)?;
        }
        if !self.failure_reason.is_empty() {
            struct_ser.serialize_field("failureReason", &self.failure_reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PriceHealth {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
            "is_healthy",
            "isHealthy",
            "has_min_sources",
            "hasMinSources",
            "deviation_ok",
            "deviationOk",
            "total_sources",
            "totalSources",
            "total_healthy_sources",
            "totalHealthySources",
            "failure_reason",
            "failureReason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
            IsHealthy,
            HasMinSources,
            DeviationOk,
            TotalSources,
            TotalHealthySources,
            FailureReason,
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
                            "denom" => Ok(GeneratedField::Denom),
                            "isHealthy" | "is_healthy" => Ok(GeneratedField::IsHealthy),
                            "hasMinSources" | "has_min_sources" => Ok(GeneratedField::HasMinSources),
                            "deviationOk" | "deviation_ok" => Ok(GeneratedField::DeviationOk),
                            "totalSources" | "total_sources" => Ok(GeneratedField::TotalSources),
                            "totalHealthySources" | "total_healthy_sources" => Ok(GeneratedField::TotalHealthySources),
                            "failureReason" | "failure_reason" => Ok(GeneratedField::FailureReason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PriceHealth;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PriceHealth")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PriceHealth, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                let mut is_healthy__ = None;
                let mut has_min_sources__ = None;
                let mut deviation_ok__ = None;
                let mut total_sources__ = None;
                let mut total_healthy_sources__ = None;
                let mut failure_reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsHealthy => {
                            if is_healthy__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isHealthy"));
                            }
                            is_healthy__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HasMinSources => {
                            if has_min_sources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hasMinSources"));
                            }
                            has_min_sources__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DeviationOk => {
                            if deviation_ok__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviationOk"));
                            }
                            deviation_ok__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalSources => {
                            if total_sources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalSources"));
                            }
                            total_sources__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TotalHealthySources => {
                            if total_healthy_sources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalHealthySources"));
                            }
                            total_healthy_sources__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::FailureReason => {
                            if failure_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("failureReason"));
                            }
                            failure_reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(PriceHealth {
                    denom: denom__.unwrap_or_default(),
                    is_healthy: is_healthy__.unwrap_or_default(),
                    has_min_sources: has_min_sources__.unwrap_or_default(),
                    deviation_ok: deviation_ok__.unwrap_or_default(),
                    total_sources: total_sources__.unwrap_or_default(),
                    total_healthy_sources: total_healthy_sources__.unwrap_or_default(),
                    failure_reason: failure_reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PriceHealth", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PricesFilter {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.asset_denom.is_empty() {
            len += 1;
        }
        if !self.base_denom.is_empty() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PricesFilter", len)?;
        if !self.asset_denom.is_empty() {
            struct_ser.serialize_field("assetDenom", &self.asset_denom)?;
        }
        if !self.base_denom.is_empty() {
            struct_ser.serialize_field("baseDenom", &self.base_denom)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PricesFilter {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "asset_denom",
            "assetDenom",
            "base_denom",
            "baseDenom",
            "height",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AssetDenom,
            BaseDenom,
            Height,
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
                            "assetDenom" | "asset_denom" => Ok(GeneratedField::AssetDenom),
                            "baseDenom" | "base_denom" => Ok(GeneratedField::BaseDenom),
                            "height" => Ok(GeneratedField::Height),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PricesFilter;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PricesFilter")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PricesFilter, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut asset_denom__ = None;
                let mut base_denom__ = None;
                let mut height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AssetDenom => {
                            if asset_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assetDenom"));
                            }
                            asset_denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BaseDenom => {
                            if base_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("baseDenom"));
                            }
                            base_denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(PricesFilter {
                    asset_denom: asset_denom__.unwrap_or_default(),
                    base_denom: base_denom__.unwrap_or_default(),
                    height: height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PricesFilter", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PythContractParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.akt_price_feed_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.PythContractParams", len)?;
        if !self.akt_price_feed_id.is_empty() {
            struct_ser.serialize_field("aktPriceFeedId", &self.akt_price_feed_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PythContractParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "akt_price_feed_id",
            "aktPriceFeedId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AktPriceFeedId,
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
                            "aktPriceFeedId" | "akt_price_feed_id" => Ok(GeneratedField::AktPriceFeedId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PythContractParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.PythContractParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PythContractParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut akt_price_feed_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AktPriceFeedId => {
                            if akt_price_feed_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("aktPriceFeedId"));
                            }
                            akt_price_feed_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(PythContractParams {
                    akt_price_feed_id: akt_price_feed_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.PythContractParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAggregatedPriceRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.denom.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryAggregatedPriceRequest", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAggregatedPriceRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
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
                            "denom" => Ok(GeneratedField::Denom),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAggregatedPriceRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.QueryAggregatedPriceRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAggregatedPriceRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAggregatedPriceRequest {
                    denom: denom__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryAggregatedPriceRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAggregatedPriceResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.aggregated_price.is_some() {
            len += 1;
        }
        if self.price_health.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryAggregatedPriceResponse", len)?;
        if let Some(v) = self.aggregated_price.as_ref() {
            struct_ser.serialize_field("aggregatedPrice", v)?;
        }
        if let Some(v) = self.price_health.as_ref() {
            struct_ser.serialize_field("priceHealth", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAggregatedPriceResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "aggregated_price",
            "aggregatedPrice",
            "price_health",
            "priceHealth",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AggregatedPrice,
            PriceHealth,
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
                            "aggregatedPrice" | "aggregated_price" => Ok(GeneratedField::AggregatedPrice),
                            "priceHealth" | "price_health" => Ok(GeneratedField::PriceHealth),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAggregatedPriceResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.QueryAggregatedPriceResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAggregatedPriceResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut aggregated_price__ = None;
                let mut price_health__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AggregatedPrice => {
                            if aggregated_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("aggregatedPrice"));
                            }
                            aggregated_price__ = map_.next_value()?;
                        }
                        GeneratedField::PriceHealth => {
                            if price_health__.is_some() {
                                return Err(serde::de::Error::duplicate_field("priceHealth"));
                            }
                            price_health__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAggregatedPriceResponse {
                    aggregated_price: aggregated_price__,
                    price_health: price_health__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryAggregatedPriceResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.oracle.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.oracle.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPriceFeedConfigRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.denom.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryPriceFeedConfigRequest", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPriceFeedConfigRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
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
                            "denom" => Ok(GeneratedField::Denom),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryPriceFeedConfigRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.QueryPriceFeedConfigRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPriceFeedConfigRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryPriceFeedConfigRequest {
                    denom: denom__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryPriceFeedConfigRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPriceFeedConfigResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.price_feed_id.is_empty() {
            len += 1;
        }
        if !self.pyth_contract_address.is_empty() {
            len += 1;
        }
        if self.enabled {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryPriceFeedConfigResponse", len)?;
        if !self.price_feed_id.is_empty() {
            struct_ser.serialize_field("priceFeedId", &self.price_feed_id)?;
        }
        if !self.pyth_contract_address.is_empty() {
            struct_ser.serialize_field("pythContractAddress", &self.pyth_contract_address)?;
        }
        if self.enabled {
            struct_ser.serialize_field("enabled", &self.enabled)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPriceFeedConfigResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "price_feed_id",
            "priceFeedId",
            "pyth_contract_address",
            "pythContractAddress",
            "enabled",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            PriceFeedId,
            PythContractAddress,
            Enabled,
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
                            "priceFeedId" | "price_feed_id" => Ok(GeneratedField::PriceFeedId),
                            "pythContractAddress" | "pyth_contract_address" => Ok(GeneratedField::PythContractAddress),
                            "enabled" => Ok(GeneratedField::Enabled),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryPriceFeedConfigResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.QueryPriceFeedConfigResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPriceFeedConfigResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut price_feed_id__ = None;
                let mut pyth_contract_address__ = None;
                let mut enabled__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::PriceFeedId => {
                            if price_feed_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("priceFeedId"));
                            }
                            price_feed_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PythContractAddress => {
                            if pyth_contract_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pythContractAddress"));
                            }
                            pyth_contract_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Enabled => {
                            if enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enabled"));
                            }
                            enabled__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryPriceFeedConfigResponse {
                    price_feed_id: price_feed_id__.unwrap_or_default(),
                    pyth_contract_address: pyth_contract_address__.unwrap_or_default(),
                    enabled: enabled__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryPriceFeedConfigResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPricesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.filters.is_some() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryPricesRequest", len)?;
        if let Some(v) = self.filters.as_ref() {
            struct_ser.serialize_field("filters", v)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPricesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "filters",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Filters,
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
                            "filters" => Ok(GeneratedField::Filters),
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
            type Value = QueryPricesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.QueryPricesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPricesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut filters__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Filters => {
                            if filters__.is_some() {
                                return Err(serde::de::Error::duplicate_field("filters"));
                            }
                            filters__ = map_.next_value()?;
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryPricesRequest {
                    filters: filters__,
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryPricesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPricesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.prices.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.oracle.v1.QueryPricesResponse", len)?;
        if !self.prices.is_empty() {
            struct_ser.serialize_field("prices", &self.prices)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPricesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "prices",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Prices,
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
                            "prices" => Ok(GeneratedField::Prices),
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
            type Value = QueryPricesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.oracle.v1.QueryPricesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPricesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut prices__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Prices => {
                            if prices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("prices"));
                            }
                            prices__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryPricesResponse {
                    prices: prices__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.oracle.v1.QueryPricesResponse", FIELDS, GeneratedVisitor)
    }
}

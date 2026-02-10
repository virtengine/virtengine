// @generated
impl serde::Serialize for BurnMintPair {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.burned.is_some() {
            len += 1;
        }
        if self.minted.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.BurnMintPair", len)?;
        if let Some(v) = self.burned.as_ref() {
            struct_ser.serialize_field("burned", v)?;
        }
        if let Some(v) = self.minted.as_ref() {
            struct_ser.serialize_field("minted", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BurnMintPair {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "burned",
            "minted",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Burned,
            Minted,
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
                            "burned" => Ok(GeneratedField::Burned),
                            "minted" => Ok(GeneratedField::Minted),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BurnMintPair;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.BurnMintPair")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<BurnMintPair, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut burned__ = None;
                let mut minted__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Burned => {
                            if burned__.is_some() {
                                return Err(serde::de::Error::duplicate_field("burned"));
                            }
                            burned__ = map_.next_value()?;
                        }
                        GeneratedField::Minted => {
                            if minted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minted"));
                            }
                            minted__ = map_.next_value()?;
                        }
                    }
                }
                Ok(BurnMintPair {
                    burned: burned__,
                    minted: minted__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.BurnMintPair", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for CoinPrice {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.coin.is_some() {
            len += 1;
        }
        if !self.price.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.CoinPrice", len)?;
        if let Some(v) = self.coin.as_ref() {
            struct_ser.serialize_field("coin", v)?;
        }
        if !self.price.is_empty() {
            struct_ser.serialize_field("price", &self.price)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for CoinPrice {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "coin",
            "price",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Coin,
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
                            "coin" => Ok(GeneratedField::Coin),
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
            type Value = CoinPrice;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.CoinPrice")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<CoinPrice, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut coin__ = None;
                let mut price__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Coin => {
                            if coin__.is_some() {
                                return Err(serde::de::Error::duplicate_field("coin"));
                            }
                            coin__ = map_.next_value()?;
                        }
                        GeneratedField::Price => {
                            if price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("price"));
                            }
                            price__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(CoinPrice {
                    coin: coin__,
                    price: price__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.CoinPrice", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for CollateralRatio {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.ratio.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.reference_price.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.CollateralRatio", len)?;
        if !self.ratio.is_empty() {
            struct_ser.serialize_field("ratio", &self.ratio)?;
        }
        if self.status != 0 {
            let v = MintStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.reference_price.is_empty() {
            struct_ser.serialize_field("referencePrice", &self.reference_price)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for CollateralRatio {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ratio",
            "status",
            "reference_price",
            "referencePrice",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Ratio,
            Status,
            ReferencePrice,
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
                            "ratio" => Ok(GeneratedField::Ratio),
                            "status" => Ok(GeneratedField::Status),
                            "referencePrice" | "reference_price" => Ok(GeneratedField::ReferencePrice),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = CollateralRatio;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.CollateralRatio")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<CollateralRatio, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut ratio__ = None;
                let mut status__ = None;
                let mut reference_price__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Ratio => {
                            if ratio__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ratio"));
                            }
                            ratio__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<MintStatus>()? as i32);
                        }
                        GeneratedField::ReferencePrice => {
                            if reference_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("referencePrice"));
                            }
                            reference_price__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(CollateralRatio {
                    ratio: ratio__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    reference_price: reference_price__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.CollateralRatio", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventLedgerRecordExecuted {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.EventLedgerRecordExecuted", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventLedgerRecordExecuted {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
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
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventLedgerRecordExecuted;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.EventLedgerRecordExecuted")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventLedgerRecordExecuted, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                    }
                }
                Ok(EventLedgerRecordExecuted {
                    id: id__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.EventLedgerRecordExecuted", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventMintStatusChange {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.previous_status != 0 {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if !self.collateral_ratio.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.EventMintStatusChange", len)?;
        if self.previous_status != 0 {
            let v = MintStatus::try_from(self.previous_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_status)))?;
            struct_ser.serialize_field("previousStatus", &v)?;
        }
        if self.new_status != 0 {
            let v = MintStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if !self.collateral_ratio.is_empty() {
            struct_ser.serialize_field("collateralRatio", &self.collateral_ratio)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventMintStatusChange {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "previous_status",
            "previousStatus",
            "new_status",
            "newStatus",
            "collateral_ratio",
            "collateralRatio",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            PreviousStatus,
            NewStatus,
            CollateralRatio,
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
                            "previousStatus" | "previous_status" => Ok(GeneratedField::PreviousStatus),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
                            "collateralRatio" | "collateral_ratio" => Ok(GeneratedField::CollateralRatio),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventMintStatusChange;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.EventMintStatusChange")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventMintStatusChange, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut previous_status__ = None;
                let mut new_status__ = None;
                let mut collateral_ratio__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::PreviousStatus => {
                            if previous_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousStatus"));
                            }
                            previous_status__ = Some(map_.next_value::<MintStatus>()? as i32);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<MintStatus>()? as i32);
                        }
                        GeneratedField::CollateralRatio => {
                            if collateral_ratio__.is_some() {
                                return Err(serde::de::Error::duplicate_field("collateralRatio"));
                            }
                            collateral_ratio__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventMintStatusChange {
                    previous_status: previous_status__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    collateral_ratio: collateral_ratio__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.EventMintStatusChange", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventVaultSeeded {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.amount.is_some() {
            len += 1;
        }
        if !self.source.is_empty() {
            len += 1;
        }
        if self.new_vault_balance.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.EventVaultSeeded", len)?;
        if let Some(v) = self.amount.as_ref() {
            struct_ser.serialize_field("amount", v)?;
        }
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if let Some(v) = self.new_vault_balance.as_ref() {
            struct_ser.serialize_field("newVaultBalance", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventVaultSeeded {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "amount",
            "source",
            "new_vault_balance",
            "newVaultBalance",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Amount,
            Source,
            NewVaultBalance,
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
                            "amount" => Ok(GeneratedField::Amount),
                            "source" => Ok(GeneratedField::Source),
                            "newVaultBalance" | "new_vault_balance" => Ok(GeneratedField::NewVaultBalance),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventVaultSeeded;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.EventVaultSeeded")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventVaultSeeded, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut amount__ = None;
                let mut source__ = None;
                let mut new_vault_balance__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = map_.next_value()?;
                        }
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewVaultBalance => {
                            if new_vault_balance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newVaultBalance"));
                            }
                            new_vault_balance__ = map_.next_value()?;
                        }
                    }
                }
                Ok(EventVaultSeeded {
                    amount: amount__,
                    source: source__.unwrap_or_default(),
                    new_vault_balance: new_vault_balance__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.EventVaultSeeded", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for GenesisLedgerPendingRecord {
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
        if self.record.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.GenesisLedgerPendingRecord", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.record.as_ref() {
            struct_ser.serialize_field("record", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GenesisLedgerPendingRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "record",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Record,
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
                            "record" => Ok(GeneratedField::Record),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GenesisLedgerPendingRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.GenesisLedgerPendingRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisLedgerPendingRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut record__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Record => {
                            if record__.is_some() {
                                return Err(serde::de::Error::duplicate_field("record"));
                            }
                            record__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisLedgerPendingRecord {
                    id: id__,
                    record: record__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.GenesisLedgerPendingRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for GenesisLedgerRecord {
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
        if self.record.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.GenesisLedgerRecord", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.record.as_ref() {
            struct_ser.serialize_field("record", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GenesisLedgerRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "record",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Record,
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
                            "record" => Ok(GeneratedField::Record),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GenesisLedgerRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.GenesisLedgerRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisLedgerRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut record__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Record => {
                            if record__.is_some() {
                                return Err(serde::de::Error::duplicate_field("record"));
                            }
                            record__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisLedgerRecord {
                    id: id__,
                    record: record__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.GenesisLedgerRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for GenesisLedgerState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.records.is_empty() {
            len += 1;
        }
        if !self.pending_records.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.GenesisLedgerState", len)?;
        if !self.records.is_empty() {
            struct_ser.serialize_field("records", &self.records)?;
        }
        if !self.pending_records.is_empty() {
            struct_ser.serialize_field("pendingRecords", &self.pending_records)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GenesisLedgerState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "records",
            "pending_records",
            "pendingRecords",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Records,
            PendingRecords,
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
                            "records" => Ok(GeneratedField::Records),
                            "pendingRecords" | "pending_records" => Ok(GeneratedField::PendingRecords),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GenesisLedgerState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.GenesisLedgerState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisLedgerState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut records__ = None;
                let mut pending_records__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Records => {
                            if records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("records"));
                            }
                            records__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PendingRecords => {
                            if pending_records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pendingRecords"));
                            }
                            pending_records__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisLedgerState {
                    records: records__.unwrap_or_default(),
                    pending_records: pending_records__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.GenesisLedgerState", FIELDS, GeneratedVisitor)
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
        if self.state.is_some() {
            len += 1;
        }
        if self.ledger.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if let Some(v) = self.state.as_ref() {
            struct_ser.serialize_field("state", v)?;
        }
        if let Some(v) = self.ledger.as_ref() {
            struct_ser.serialize_field("ledger", v)?;
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
            "state",
            "ledger",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            State,
            Ledger,
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
                            "state" => Ok(GeneratedField::State),
                            "ledger" => Ok(GeneratedField::Ledger),
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
                formatter.write_str("struct virtengine.bme.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut state__ = None;
                let mut ledger__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = map_.next_value()?;
                        }
                        GeneratedField::Ledger => {
                            if ledger__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ledger"));
                            }
                            ledger__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    state: state__,
                    ledger: ledger__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for GenesisVaultState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.total_burned.is_empty() {
            len += 1;
        }
        if !self.total_minted.is_empty() {
            len += 1;
        }
        if !self.remint_credits.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.GenesisVaultState", len)?;
        if !self.total_burned.is_empty() {
            struct_ser.serialize_field("totalBurned", &self.total_burned)?;
        }
        if !self.total_minted.is_empty() {
            struct_ser.serialize_field("totalMinted", &self.total_minted)?;
        }
        if !self.remint_credits.is_empty() {
            struct_ser.serialize_field("remintCredits", &self.remint_credits)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GenesisVaultState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "total_burned",
            "totalBurned",
            "total_minted",
            "totalMinted",
            "remint_credits",
            "remintCredits",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TotalBurned,
            TotalMinted,
            RemintCredits,
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
                            "totalBurned" | "total_burned" => Ok(GeneratedField::TotalBurned),
                            "totalMinted" | "total_minted" => Ok(GeneratedField::TotalMinted),
                            "remintCredits" | "remint_credits" => Ok(GeneratedField::RemintCredits),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GenesisVaultState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.GenesisVaultState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisVaultState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut total_burned__ = None;
                let mut total_minted__ = None;
                let mut remint_credits__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TotalBurned => {
                            if total_burned__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalBurned"));
                            }
                            total_burned__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalMinted => {
                            if total_minted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalMinted"));
                            }
                            total_minted__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RemintCredits => {
                            if remint_credits__.is_some() {
                                return Err(serde::de::Error::duplicate_field("remintCredits"));
                            }
                            remint_credits__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisVaultState {
                    total_burned: total_burned__.unwrap_or_default(),
                    total_minted: total_minted__.unwrap_or_default(),
                    remint_credits: remint_credits__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.GenesisVaultState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LedgerId {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.height != 0 {
            len += 1;
        }
        if self.sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.LedgerID", len)?;
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        if self.sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sequence", ToString::to_string(&self.sequence).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LedgerId {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "height",
            "sequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Height,
            Sequence,
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
                            "height" => Ok(GeneratedField::Height),
                            "sequence" => Ok(GeneratedField::Sequence),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LedgerId;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.LedgerID")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LedgerId, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut height__ = None;
                let mut sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Sequence => {
                            if sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sequence"));
                            }
                            sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(LedgerId {
                    height: height__.unwrap_or_default(),
                    sequence: sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.LedgerID", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LedgerPendingRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.owner.is_empty() {
            len += 1;
        }
        if !self.to.is_empty() {
            len += 1;
        }
        if self.coins_to_burn.is_some() {
            len += 1;
        }
        if !self.denom_to_mint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.LedgerPendingRecord", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.to.is_empty() {
            struct_ser.serialize_field("to", &self.to)?;
        }
        if let Some(v) = self.coins_to_burn.as_ref() {
            struct_ser.serialize_field("coinsToBurn", v)?;
        }
        if !self.denom_to_mint.is_empty() {
            struct_ser.serialize_field("denomToMint", &self.denom_to_mint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LedgerPendingRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "to",
            "coins_to_burn",
            "coinsToBurn",
            "denom_to_mint",
            "denomToMint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            To,
            CoinsToBurn,
            DenomToMint,
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
                            "owner" => Ok(GeneratedField::Owner),
                            "to" => Ok(GeneratedField::To),
                            "coinsToBurn" | "coins_to_burn" => Ok(GeneratedField::CoinsToBurn),
                            "denomToMint" | "denom_to_mint" => Ok(GeneratedField::DenomToMint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LedgerPendingRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.LedgerPendingRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LedgerPendingRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut to__ = None;
                let mut coins_to_burn__ = None;
                let mut denom_to_mint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::To => {
                            if to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("to"));
                            }
                            to__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CoinsToBurn => {
                            if coins_to_burn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("coinsToBurn"));
                            }
                            coins_to_burn__ = map_.next_value()?;
                        }
                        GeneratedField::DenomToMint => {
                            if denom_to_mint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denomToMint"));
                            }
                            denom_to_mint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(LedgerPendingRecord {
                    owner: owner__.unwrap_or_default(),
                    to: to__.unwrap_or_default(),
                    coins_to_burn: coins_to_burn__,
                    denom_to_mint: denom_to_mint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.LedgerPendingRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LedgerRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.burned_from.is_empty() {
            len += 1;
        }
        if !self.minted_to.is_empty() {
            len += 1;
        }
        if !self.burner.is_empty() {
            len += 1;
        }
        if !self.minter.is_empty() {
            len += 1;
        }
        if self.burned.is_some() {
            len += 1;
        }
        if self.minted.is_some() {
            len += 1;
        }
        if self.remint_credit_issued.is_some() {
            len += 1;
        }
        if self.remint_credit_accrued.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.LedgerRecord", len)?;
        if !self.burned_from.is_empty() {
            struct_ser.serialize_field("burnedFrom", &self.burned_from)?;
        }
        if !self.minted_to.is_empty() {
            struct_ser.serialize_field("mintedTo", &self.minted_to)?;
        }
        if !self.burner.is_empty() {
            struct_ser.serialize_field("burner", &self.burner)?;
        }
        if !self.minter.is_empty() {
            struct_ser.serialize_field("minter", &self.minter)?;
        }
        if let Some(v) = self.burned.as_ref() {
            struct_ser.serialize_field("burned", v)?;
        }
        if let Some(v) = self.minted.as_ref() {
            struct_ser.serialize_field("minted", v)?;
        }
        if let Some(v) = self.remint_credit_issued.as_ref() {
            struct_ser.serialize_field("remintCreditIssued", v)?;
        }
        if let Some(v) = self.remint_credit_accrued.as_ref() {
            struct_ser.serialize_field("remintCreditAccrued", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LedgerRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "burned_from",
            "burnedFrom",
            "minted_to",
            "mintedTo",
            "burner",
            "minter",
            "burned",
            "minted",
            "remint_credit_issued",
            "remintCreditIssued",
            "remint_credit_accrued",
            "remintCreditAccrued",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            BurnedFrom,
            MintedTo,
            Burner,
            Minter,
            Burned,
            Minted,
            RemintCreditIssued,
            RemintCreditAccrued,
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
                            "burnedFrom" | "burned_from" => Ok(GeneratedField::BurnedFrom),
                            "mintedTo" | "minted_to" => Ok(GeneratedField::MintedTo),
                            "burner" => Ok(GeneratedField::Burner),
                            "minter" => Ok(GeneratedField::Minter),
                            "burned" => Ok(GeneratedField::Burned),
                            "minted" => Ok(GeneratedField::Minted),
                            "remintCreditIssued" | "remint_credit_issued" => Ok(GeneratedField::RemintCreditIssued),
                            "remintCreditAccrued" | "remint_credit_accrued" => Ok(GeneratedField::RemintCreditAccrued),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LedgerRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.LedgerRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LedgerRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut burned_from__ = None;
                let mut minted_to__ = None;
                let mut burner__ = None;
                let mut minter__ = None;
                let mut burned__ = None;
                let mut minted__ = None;
                let mut remint_credit_issued__ = None;
                let mut remint_credit_accrued__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::BurnedFrom => {
                            if burned_from__.is_some() {
                                return Err(serde::de::Error::duplicate_field("burnedFrom"));
                            }
                            burned_from__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MintedTo => {
                            if minted_to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mintedTo"));
                            }
                            minted_to__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Burner => {
                            if burner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("burner"));
                            }
                            burner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Minter => {
                            if minter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minter"));
                            }
                            minter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Burned => {
                            if burned__.is_some() {
                                return Err(serde::de::Error::duplicate_field("burned"));
                            }
                            burned__ = map_.next_value()?;
                        }
                        GeneratedField::Minted => {
                            if minted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minted"));
                            }
                            minted__ = map_.next_value()?;
                        }
                        GeneratedField::RemintCreditIssued => {
                            if remint_credit_issued__.is_some() {
                                return Err(serde::de::Error::duplicate_field("remintCreditIssued"));
                            }
                            remint_credit_issued__ = map_.next_value()?;
                        }
                        GeneratedField::RemintCreditAccrued => {
                            if remint_credit_accrued__.is_some() {
                                return Err(serde::de::Error::duplicate_field("remintCreditAccrued"));
                            }
                            remint_credit_accrued__ = map_.next_value()?;
                        }
                    }
                }
                Ok(LedgerRecord {
                    burned_from: burned_from__.unwrap_or_default(),
                    minted_to: minted_to__.unwrap_or_default(),
                    burner: burner__.unwrap_or_default(),
                    minter: minter__.unwrap_or_default(),
                    burned: burned__,
                    minted: minted__,
                    remint_credit_issued: remint_credit_issued__,
                    remint_credit_accrued: remint_credit_accrued__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.LedgerRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LedgerRecordId {
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
        if !self.to_denom.is_empty() {
            len += 1;
        }
        if !self.source.is_empty() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        if self.sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.LedgerRecordID", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if !self.to_denom.is_empty() {
            struct_ser.serialize_field("toDenom", &self.to_denom)?;
        }
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        if self.sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sequence", ToString::to_string(&self.sequence).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for LedgerRecordId {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
            "to_denom",
            "toDenom",
            "source",
            "height",
            "sequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
            ToDenom,
            Source,
            Height,
            Sequence,
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
                            "toDenom" | "to_denom" => Ok(GeneratedField::ToDenom),
                            "source" => Ok(GeneratedField::Source),
                            "height" => Ok(GeneratedField::Height),
                            "sequence" => Ok(GeneratedField::Sequence),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LedgerRecordId;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.LedgerRecordID")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<LedgerRecordId, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                let mut to_denom__ = None;
                let mut source__ = None;
                let mut height__ = None;
                let mut sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ToDenom => {
                            if to_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("toDenom"));
                            }
                            to_denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Sequence => {
                            if sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sequence"));
                            }
                            sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(LedgerRecordId {
                    denom: denom__.unwrap_or_default(),
                    to_denom: to_denom__.unwrap_or_default(),
                    source: source__.unwrap_or_default(),
                    height: height__.unwrap_or_default(),
                    sequence: sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.LedgerRecordID", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for LedgerRecordStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Invalid => "ledger_record_status_invalid",
            Self::Pending => "ledger_record_status_pending",
            Self::Executed => "ledger_record_status_executed",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for LedgerRecordStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ledger_record_status_invalid",
            "ledger_record_status_pending",
            "ledger_record_status_executed",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = LedgerRecordStatus;

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
                    "ledger_record_status_invalid" => Ok(LedgerRecordStatus::Invalid),
                    "ledger_record_status_pending" => Ok(LedgerRecordStatus::Pending),
                    "ledger_record_status_executed" => Ok(LedgerRecordStatus::Executed),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for MintEpoch {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.next_epoch != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MintEpoch", len)?;
        if self.next_epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nextEpoch", ToString::to_string(&self.next_epoch).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MintEpoch {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "next_epoch",
            "nextEpoch",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            NextEpoch,
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
                            "nextEpoch" | "next_epoch" => Ok(GeneratedField::NextEpoch),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MintEpoch;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MintEpoch")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MintEpoch, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut next_epoch__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::NextEpoch => {
                            if next_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nextEpoch"));
                            }
                            next_epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MintEpoch {
                    next_epoch: next_epoch__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MintEpoch", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MintStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "mint_status_unspecified",
            Self::Healthy => "mint_status_healthy",
            Self::Warning => "mint_status_warning",
            Self::HaltCr => "mint_status_halt_cr",
            Self::HaltOracle => "mint_status_halt_oracle",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for MintStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "mint_status_unspecified",
            "mint_status_healthy",
            "mint_status_warning",
            "mint_status_halt_cr",
            "mint_status_halt_oracle",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MintStatus;

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
                    "mint_status_unspecified" => Ok(MintStatus::Unspecified),
                    "mint_status_healthy" => Ok(MintStatus::Healthy),
                    "mint_status_warning" => Ok(MintStatus::Warning),
                    "mint_status_halt_cr" => Ok(MintStatus::HaltCr),
                    "mint_status_halt_oracle" => Ok(MintStatus::HaltOracle),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for MsgBurnAct {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.owner.is_empty() {
            len += 1;
        }
        if !self.to.is_empty() {
            len += 1;
        }
        if self.coins_to_burn.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgBurnACT", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.to.is_empty() {
            struct_ser.serialize_field("to", &self.to)?;
        }
        if let Some(v) = self.coins_to_burn.as_ref() {
            struct_ser.serialize_field("coinsToBurn", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgBurnAct {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "to",
            "coins_to_burn",
            "coinsToBurn",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            To,
            CoinsToBurn,
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
                            "owner" => Ok(GeneratedField::Owner),
                            "to" => Ok(GeneratedField::To),
                            "coinsToBurn" | "coins_to_burn" => Ok(GeneratedField::CoinsToBurn),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgBurnAct;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgBurnACT")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgBurnAct, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut to__ = None;
                let mut coins_to_burn__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::To => {
                            if to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("to"));
                            }
                            to__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CoinsToBurn => {
                            if coins_to_burn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("coinsToBurn"));
                            }
                            coins_to_burn__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgBurnAct {
                    owner: owner__.unwrap_or_default(),
                    to: to__.unwrap_or_default(),
                    coins_to_burn: coins_to_burn__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgBurnACT", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgBurnActResponse {
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
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgBurnACTResponse", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.status != 0 {
            let v = LedgerRecordStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgBurnActResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Status,
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
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgBurnActResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgBurnACTResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgBurnActResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<LedgerRecordStatus>()? as i32);
                        }
                    }
                }
                Ok(MsgBurnActResponse {
                    id: id__,
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgBurnACTResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgBurnMint {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.owner.is_empty() {
            len += 1;
        }
        if !self.to.is_empty() {
            len += 1;
        }
        if self.coins_to_burn.is_some() {
            len += 1;
        }
        if !self.denom_to_mint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgBurnMint", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.to.is_empty() {
            struct_ser.serialize_field("to", &self.to)?;
        }
        if let Some(v) = self.coins_to_burn.as_ref() {
            struct_ser.serialize_field("coinsToBurn", v)?;
        }
        if !self.denom_to_mint.is_empty() {
            struct_ser.serialize_field("denomToMint", &self.denom_to_mint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgBurnMint {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "to",
            "coins_to_burn",
            "coinsToBurn",
            "denom_to_mint",
            "denomToMint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            To,
            CoinsToBurn,
            DenomToMint,
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
                            "owner" => Ok(GeneratedField::Owner),
                            "to" => Ok(GeneratedField::To),
                            "coinsToBurn" | "coins_to_burn" => Ok(GeneratedField::CoinsToBurn),
                            "denomToMint" | "denom_to_mint" => Ok(GeneratedField::DenomToMint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgBurnMint;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgBurnMint")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgBurnMint, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut to__ = None;
                let mut coins_to_burn__ = None;
                let mut denom_to_mint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::To => {
                            if to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("to"));
                            }
                            to__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CoinsToBurn => {
                            if coins_to_burn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("coinsToBurn"));
                            }
                            coins_to_burn__ = map_.next_value()?;
                        }
                        GeneratedField::DenomToMint => {
                            if denom_to_mint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denomToMint"));
                            }
                            denom_to_mint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgBurnMint {
                    owner: owner__.unwrap_or_default(),
                    to: to__.unwrap_or_default(),
                    coins_to_burn: coins_to_burn__,
                    denom_to_mint: denom_to_mint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgBurnMint", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgBurnMintResponse {
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
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgBurnMintResponse", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.status != 0 {
            let v = LedgerRecordStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgBurnMintResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Status,
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
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgBurnMintResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgBurnMintResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgBurnMintResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<LedgerRecordStatus>()? as i32);
                        }
                    }
                }
                Ok(MsgBurnMintResponse {
                    id: id__,
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgBurnMintResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgMintAct {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.owner.is_empty() {
            len += 1;
        }
        if !self.to.is_empty() {
            len += 1;
        }
        if self.coins_to_burn.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgMintACT", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.to.is_empty() {
            struct_ser.serialize_field("to", &self.to)?;
        }
        if let Some(v) = self.coins_to_burn.as_ref() {
            struct_ser.serialize_field("coinsToBurn", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgMintAct {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "to",
            "coins_to_burn",
            "coinsToBurn",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            To,
            CoinsToBurn,
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
                            "owner" => Ok(GeneratedField::Owner),
                            "to" => Ok(GeneratedField::To),
                            "coinsToBurn" | "coins_to_burn" => Ok(GeneratedField::CoinsToBurn),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgMintAct;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgMintACT")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgMintAct, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut to__ = None;
                let mut coins_to_burn__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::To => {
                            if to__.is_some() {
                                return Err(serde::de::Error::duplicate_field("to"));
                            }
                            to__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CoinsToBurn => {
                            if coins_to_burn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("coinsToBurn"));
                            }
                            coins_to_burn__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgMintAct {
                    owner: owner__.unwrap_or_default(),
                    to: to__.unwrap_or_default(),
                    coins_to_burn: coins_to_burn__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgMintACT", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgMintActResponse {
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
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgMintACTResponse", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.status != 0 {
            let v = LedgerRecordStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgMintActResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            Status,
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
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgMintActResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgMintACTResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgMintActResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<LedgerRecordStatus>()? as i32);
                        }
                    }
                }
                Ok(MsgMintActResponse {
                    id: id__,
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgMintACTResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSeedVault {
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
        if self.amount.is_some() {
            len += 1;
        }
        if !self.source.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgSeedVault", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.amount.as_ref() {
            struct_ser.serialize_field("amount", v)?;
        }
        if !self.source.is_empty() {
            struct_ser.serialize_field("source", &self.source)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSeedVault {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "amount",
            "source",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            Amount,
            Source,
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
                            "amount" => Ok(GeneratedField::Amount),
                            "source" => Ok(GeneratedField::Source),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSeedVault;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgSeedVault")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSeedVault, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut amount__ = None;
                let mut source__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = map_.next_value()?;
                        }
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSeedVault {
                    authority: authority__.unwrap_or_default(),
                    amount: amount__,
                    source: source__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgSeedVault", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSeedVaultResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.vault_akt.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgSeedVaultResponse", len)?;
        if !self.vault_akt.is_empty() {
            struct_ser.serialize_field("vaultAkt", &self.vault_akt)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSeedVaultResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "vault_akt",
            "vaultAkt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            VaultAkt,
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
                            "vaultAkt" | "vault_akt" => Ok(GeneratedField::VaultAkt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSeedVaultResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.MsgSeedVaultResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSeedVaultResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut vault_akt__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::VaultAkt => {
                            if vault_akt__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vaultAkt"));
                            }
                            vault_akt__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSeedVaultResponse {
                    vault_akt: vault_akt__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.MsgSeedVaultResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.bme.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.bme.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.bme.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.bme.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.bme.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if self.circuit_breaker_warn_threshold != 0 {
            len += 1;
        }
        if self.circuit_breaker_halt_threshold != 0 {
            len += 1;
        }
        if self.min_epoch_blocks != 0 {
            len += 1;
        }
        if self.epoch_blocks_backoff != 0 {
            len += 1;
        }
        if self.mint_spread_bps != 0 {
            len += 1;
        }
        if self.settle_spread_bps != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.Params", len)?;
        if self.circuit_breaker_warn_threshold != 0 {
            struct_ser.serialize_field("circuitBreakerWarnThreshold", &self.circuit_breaker_warn_threshold)?;
        }
        if self.circuit_breaker_halt_threshold != 0 {
            struct_ser.serialize_field("circuitBreakerHaltThreshold", &self.circuit_breaker_halt_threshold)?;
        }
        if self.min_epoch_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("minEpochBlocks", ToString::to_string(&self.min_epoch_blocks).as_str())?;
        }
        if self.epoch_blocks_backoff != 0 {
            struct_ser.serialize_field("epochBlocksBackoff", &self.epoch_blocks_backoff)?;
        }
        if self.mint_spread_bps != 0 {
            struct_ser.serialize_field("mintSpreadBps", &self.mint_spread_bps)?;
        }
        if self.settle_spread_bps != 0 {
            struct_ser.serialize_field("settleSpreadBps", &self.settle_spread_bps)?;
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
            "circuit_breaker_warn_threshold",
            "circuitBreakerWarnThreshold",
            "circuit_breaker_halt_threshold",
            "circuitBreakerHaltThreshold",
            "min_epoch_blocks",
            "minEpochBlocks",
            "epoch_blocks_backoff",
            "epochBlocksBackoff",
            "mint_spread_bps",
            "mintSpreadBps",
            "settle_spread_bps",
            "settleSpreadBps",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CircuitBreakerWarnThreshold,
            CircuitBreakerHaltThreshold,
            MinEpochBlocks,
            EpochBlocksBackoff,
            MintSpreadBps,
            SettleSpreadBps,
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
                            "circuitBreakerWarnThreshold" | "circuit_breaker_warn_threshold" => Ok(GeneratedField::CircuitBreakerWarnThreshold),
                            "circuitBreakerHaltThreshold" | "circuit_breaker_halt_threshold" => Ok(GeneratedField::CircuitBreakerHaltThreshold),
                            "minEpochBlocks" | "min_epoch_blocks" => Ok(GeneratedField::MinEpochBlocks),
                            "epochBlocksBackoff" | "epoch_blocks_backoff" => Ok(GeneratedField::EpochBlocksBackoff),
                            "mintSpreadBps" | "mint_spread_bps" => Ok(GeneratedField::MintSpreadBps),
                            "settleSpreadBps" | "settle_spread_bps" => Ok(GeneratedField::SettleSpreadBps),
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
                formatter.write_str("struct virtengine.bme.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut circuit_breaker_warn_threshold__ = None;
                let mut circuit_breaker_halt_threshold__ = None;
                let mut min_epoch_blocks__ = None;
                let mut epoch_blocks_backoff__ = None;
                let mut mint_spread_bps__ = None;
                let mut settle_spread_bps__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CircuitBreakerWarnThreshold => {
                            if circuit_breaker_warn_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("circuitBreakerWarnThreshold"));
                            }
                            circuit_breaker_warn_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CircuitBreakerHaltThreshold => {
                            if circuit_breaker_halt_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("circuitBreakerHaltThreshold"));
                            }
                            circuit_breaker_halt_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinEpochBlocks => {
                            if min_epoch_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minEpochBlocks"));
                            }
                            min_epoch_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EpochBlocksBackoff => {
                            if epoch_blocks_backoff__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochBlocksBackoff"));
                            }
                            epoch_blocks_backoff__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MintSpreadBps => {
                            if mint_spread_bps__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mintSpreadBps"));
                            }
                            mint_spread_bps__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SettleSpreadBps => {
                            if settle_spread_bps__.is_some() {
                                return Err(serde::de::Error::duplicate_field("settleSpreadBps"));
                            }
                            settle_spread_bps__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Params {
                    circuit_breaker_warn_threshold: circuit_breaker_warn_threshold__.unwrap_or_default(),
                    circuit_breaker_halt_threshold: circuit_breaker_halt_threshold__.unwrap_or_default(),
                    min_epoch_blocks: min_epoch_blocks__.unwrap_or_default(),
                    epoch_blocks_backoff: epoch_blocks_backoff__.unwrap_or_default(),
                    mint_spread_bps: mint_spread_bps__.unwrap_or_default(),
                    settle_spread_bps: settle_spread_bps__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.Params", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.bme.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.bme.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.bme.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.bme.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.bme.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryStatusRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.bme.v1.QueryStatusRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryStatusRequest {
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
            type Value = QueryStatusRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.QueryStatusRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryStatusRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryStatusRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.QueryStatusRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryStatusResponse {
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
        if !self.collateral_ratio.is_empty() {
            len += 1;
        }
        if !self.warn_threshold.is_empty() {
            len += 1;
        }
        if !self.halt_threshold.is_empty() {
            len += 1;
        }
        if self.mints_allowed {
            len += 1;
        }
        if self.refunds_allowed {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.QueryStatusResponse", len)?;
        if self.status != 0 {
            let v = MintStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.collateral_ratio.is_empty() {
            struct_ser.serialize_field("collateralRatio", &self.collateral_ratio)?;
        }
        if !self.warn_threshold.is_empty() {
            struct_ser.serialize_field("warnThreshold", &self.warn_threshold)?;
        }
        if !self.halt_threshold.is_empty() {
            struct_ser.serialize_field("haltThreshold", &self.halt_threshold)?;
        }
        if self.mints_allowed {
            struct_ser.serialize_field("mintsAllowed", &self.mints_allowed)?;
        }
        if self.refunds_allowed {
            struct_ser.serialize_field("refundsAllowed", &self.refunds_allowed)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryStatusResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "status",
            "collateral_ratio",
            "collateralRatio",
            "warn_threshold",
            "warnThreshold",
            "halt_threshold",
            "haltThreshold",
            "mints_allowed",
            "mintsAllowed",
            "refunds_allowed",
            "refundsAllowed",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Status,
            CollateralRatio,
            WarnThreshold,
            HaltThreshold,
            MintsAllowed,
            RefundsAllowed,
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
                            "collateralRatio" | "collateral_ratio" => Ok(GeneratedField::CollateralRatio),
                            "warnThreshold" | "warn_threshold" => Ok(GeneratedField::WarnThreshold),
                            "haltThreshold" | "halt_threshold" => Ok(GeneratedField::HaltThreshold),
                            "mintsAllowed" | "mints_allowed" => Ok(GeneratedField::MintsAllowed),
                            "refundsAllowed" | "refunds_allowed" => Ok(GeneratedField::RefundsAllowed),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.QueryStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut status__ = None;
                let mut collateral_ratio__ = None;
                let mut warn_threshold__ = None;
                let mut halt_threshold__ = None;
                let mut mints_allowed__ = None;
                let mut refunds_allowed__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<MintStatus>()? as i32);
                        }
                        GeneratedField::CollateralRatio => {
                            if collateral_ratio__.is_some() {
                                return Err(serde::de::Error::duplicate_field("collateralRatio"));
                            }
                            collateral_ratio__ = Some(map_.next_value()?);
                        }
                        GeneratedField::WarnThreshold => {
                            if warn_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("warnThreshold"));
                            }
                            warn_threshold__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HaltThreshold => {
                            if halt_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("haltThreshold"));
                            }
                            halt_threshold__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MintsAllowed => {
                            if mints_allowed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mintsAllowed"));
                            }
                            mints_allowed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RefundsAllowed => {
                            if refunds_allowed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("refundsAllowed"));
                            }
                            refunds_allowed__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryStatusResponse {
                    status: status__.unwrap_or_default(),
                    collateral_ratio: collateral_ratio__.unwrap_or_default(),
                    warn_threshold: warn_threshold__.unwrap_or_default(),
                    halt_threshold: halt_threshold__.unwrap_or_default(),
                    mints_allowed: mints_allowed__.unwrap_or_default(),
                    refunds_allowed: refunds_allowed__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.QueryStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryVaultStateRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.bme.v1.QueryVaultStateRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryVaultStateRequest {
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
            type Value = QueryVaultStateRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.QueryVaultStateRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryVaultStateRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryVaultStateRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.QueryVaultStateRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryVaultStateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.vault_state.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.QueryVaultStateResponse", len)?;
        if let Some(v) = self.vault_state.as_ref() {
            struct_ser.serialize_field("vaultState", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryVaultStateResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "vault_state",
            "vaultState",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            VaultState,
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
                            "vaultState" | "vault_state" => Ok(GeneratedField::VaultState),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryVaultStateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.QueryVaultStateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryVaultStateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut vault_state__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::VaultState => {
                            if vault_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vaultState"));
                            }
                            vault_state__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryVaultStateResponse {
                    vault_state: vault_state__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.QueryVaultStateResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for State {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.balances.is_empty() {
            len += 1;
        }
        if !self.total_burned.is_empty() {
            len += 1;
        }
        if !self.total_minted.is_empty() {
            len += 1;
        }
        if !self.remint_credits.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.State", len)?;
        if !self.balances.is_empty() {
            struct_ser.serialize_field("balances", &self.balances)?;
        }
        if !self.total_burned.is_empty() {
            struct_ser.serialize_field("totalBurned", &self.total_burned)?;
        }
        if !self.total_minted.is_empty() {
            struct_ser.serialize_field("totalMinted", &self.total_minted)?;
        }
        if !self.remint_credits.is_empty() {
            struct_ser.serialize_field("remintCredits", &self.remint_credits)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for State {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "balances",
            "total_burned",
            "totalBurned",
            "total_minted",
            "totalMinted",
            "remint_credits",
            "remintCredits",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Balances,
            TotalBurned,
            TotalMinted,
            RemintCredits,
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
                            "balances" => Ok(GeneratedField::Balances),
                            "totalBurned" | "total_burned" => Ok(GeneratedField::TotalBurned),
                            "totalMinted" | "total_minted" => Ok(GeneratedField::TotalMinted),
                            "remintCredits" | "remint_credits" => Ok(GeneratedField::RemintCredits),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = State;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.bme.v1.State")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<State, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut balances__ = None;
                let mut total_burned__ = None;
                let mut total_minted__ = None;
                let mut remint_credits__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Balances => {
                            if balances__.is_some() {
                                return Err(serde::de::Error::duplicate_field("balances"));
                            }
                            balances__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalBurned => {
                            if total_burned__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalBurned"));
                            }
                            total_burned__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalMinted => {
                            if total_minted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalMinted"));
                            }
                            total_minted__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RemintCredits => {
                            if remint_credits__.is_some() {
                                return Err(serde::de::Error::duplicate_field("remintCredits"));
                            }
                            remint_credits__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(State {
                    balances: balances__.unwrap_or_default(),
                    total_burned: total_burned__.unwrap_or_default(),
                    total_minted: total_minted__.unwrap_or_default(),
                    remint_credits: remint_credits__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.State", FIELDS, GeneratedVisitor)
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
        if self.status != 0 {
            len += 1;
        }
        if self.previous_status != 0 {
            len += 1;
        }
        if self.epoch_height_diff != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.bme.v1.Status", len)?;
        if self.status != 0 {
            let v = MintStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.previous_status != 0 {
            let v = MintStatus::try_from(self.previous_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_status)))?;
            struct_ser.serialize_field("previousStatus", &v)?;
        }
        if self.epoch_height_diff != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochHeightDiff", ToString::to_string(&self.epoch_height_diff).as_str())?;
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
            "status",
            "previous_status",
            "previousStatus",
            "epoch_height_diff",
            "epochHeightDiff",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Status,
            PreviousStatus,
            EpochHeightDiff,
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
                            "previousStatus" | "previous_status" => Ok(GeneratedField::PreviousStatus),
                            "epochHeightDiff" | "epoch_height_diff" => Ok(GeneratedField::EpochHeightDiff),
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
                formatter.write_str("struct virtengine.bme.v1.Status")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Status, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut status__ = None;
                let mut previous_status__ = None;
                let mut epoch_height_diff__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<MintStatus>()? as i32);
                        }
                        GeneratedField::PreviousStatus => {
                            if previous_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousStatus"));
                            }
                            previous_status__ = Some(map_.next_value::<MintStatus>()? as i32);
                        }
                        GeneratedField::EpochHeightDiff => {
                            if epoch_height_diff__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochHeightDiff"));
                            }
                            epoch_height_diff__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Status {
                    status: status__.unwrap_or_default(),
                    previous_status: previous_status__.unwrap_or_default(),
                    epoch_height_diff: epoch_height_diff__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.bme.v1.Status", FIELDS, GeneratedVisitor)
    }
}

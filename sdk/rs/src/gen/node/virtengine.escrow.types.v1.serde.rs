// @generated
impl serde::Serialize for Account {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.types.v1.Account", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.state.as_ref() {
            struct_ser.serialize_field("state", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Account {
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
            type Value = Account;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.types.v1.Account")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Account, V::Error>
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
                Ok(Account {
                    id: id__,
                    state: state__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.types.v1.Account", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AccountState {
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
        if self.state != 0 {
            len += 1;
        }
        if !self.transferred.is_empty() {
            len += 1;
        }
        if self.settled_at != 0 {
            len += 1;
        }
        if !self.funds.is_empty() {
            len += 1;
        }
        if !self.deposits.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.types.v1.AccountState", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if self.state != 0 {
            let v = State::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.transferred.is_empty() {
            struct_ser.serialize_field("transferred", &self.transferred)?;
        }
        if self.settled_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("settledAt", ToString::to_string(&self.settled_at).as_str())?;
        }
        if !self.funds.is_empty() {
            struct_ser.serialize_field("funds", &self.funds)?;
        }
        if !self.deposits.is_empty() {
            struct_ser.serialize_field("deposits", &self.deposits)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AccountState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "state",
            "transferred",
            "settled_at",
            "settledAt",
            "funds",
            "deposits",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            State,
            Transferred,
            SettledAt,
            Funds,
            Deposits,
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
                            "state" => Ok(GeneratedField::State),
                            "transferred" => Ok(GeneratedField::Transferred),
                            "settledAt" | "settled_at" => Ok(GeneratedField::SettledAt),
                            "funds" => Ok(GeneratedField::Funds),
                            "deposits" => Ok(GeneratedField::Deposits),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AccountState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.types.v1.AccountState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AccountState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut state__ = None;
                let mut transferred__ = None;
                let mut settled_at__ = None;
                let mut funds__ = None;
                let mut deposits__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<State>()? as i32);
                        }
                        GeneratedField::Transferred => {
                            if transferred__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transferred"));
                            }
                            transferred__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SettledAt => {
                            if settled_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("settledAt"));
                            }
                            settled_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Funds => {
                            if funds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("funds"));
                            }
                            funds__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Deposits => {
                            if deposits__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deposits"));
                            }
                            deposits__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(AccountState {
                    owner: owner__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    transferred: transferred__.unwrap_or_default(),
                    settled_at: settled_at__.unwrap_or_default(),
                    funds: funds__.unwrap_or_default(),
                    deposits: deposits__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.types.v1.AccountState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Balance {
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
        if !self.amount.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.types.v1.Balance", len)?;
        if !self.denom.is_empty() {
            struct_ser.serialize_field("denom", &self.denom)?;
        }
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Balance {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "denom",
            "amount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Denom,
            Amount,
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
                            "amount" => Ok(GeneratedField::Amount),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Balance;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.types.v1.Balance")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Balance, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut denom__ = None;
                let mut amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Denom => {
                            if denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denom"));
                            }
                            denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Balance {
                    denom: denom__.unwrap_or_default(),
                    amount: amount__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.types.v1.Balance", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Depositor {
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
        if self.height != 0 {
            len += 1;
        }
        if self.source != 0 {
            len += 1;
        }
        if self.balance.is_some() {
            len += 1;
        }
        if self.direct {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.types.v1.Depositor", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        if self.source != 0 {
            let v = super::super::super::base::deposit::v1::Source::try_from(self.source)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.source)))?;
            struct_ser.serialize_field("source", &v)?;
        }
        if let Some(v) = self.balance.as_ref() {
            struct_ser.serialize_field("balance", v)?;
        }
        if self.direct {
            struct_ser.serialize_field("direct", &self.direct)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Depositor {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "height",
            "source",
            "balance",
            "direct",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            Height,
            Source,
            Balance,
            Direct,
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
                            "height" => Ok(GeneratedField::Height),
                            "source" => Ok(GeneratedField::Source),
                            "balance" => Ok(GeneratedField::Balance),
                            "direct" => Ok(GeneratedField::Direct),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Depositor;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.types.v1.Depositor")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Depositor, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut height__ = None;
                let mut source__ = None;
                let mut balance__ = None;
                let mut direct__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Source => {
                            if source__.is_some() {
                                return Err(serde::de::Error::duplicate_field("source"));
                            }
                            source__ = Some(map_.next_value::<super::super::super::base::deposit::v1::Source>()? as i32);
                        }
                        GeneratedField::Balance => {
                            if balance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("balance"));
                            }
                            balance__ = map_.next_value()?;
                        }
                        GeneratedField::Direct => {
                            if direct__.is_some() {
                                return Err(serde::de::Error::duplicate_field("direct"));
                            }
                            direct__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Depositor {
                    owner: owner__.unwrap_or_default(),
                    height: height__.unwrap_or_default(),
                    source: source__.unwrap_or_default(),
                    balance: balance__,
                    direct: direct__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.types.v1.Depositor", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Payment {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.types.v1.Payment", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.state.as_ref() {
            struct_ser.serialize_field("state", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Payment {
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
            type Value = Payment;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.types.v1.Payment")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Payment, V::Error>
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
                Ok(Payment {
                    id: id__,
                    state: state__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.types.v1.Payment", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PaymentState {
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
        if self.state != 0 {
            len += 1;
        }
        if self.rate.is_some() {
            len += 1;
        }
        if self.balance.is_some() {
            len += 1;
        }
        if self.unsettled.is_some() {
            len += 1;
        }
        if self.withdrawn.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.types.v1.PaymentState", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if self.state != 0 {
            let v = State::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if let Some(v) = self.rate.as_ref() {
            struct_ser.serialize_field("rate", v)?;
        }
        if let Some(v) = self.balance.as_ref() {
            struct_ser.serialize_field("balance", v)?;
        }
        if let Some(v) = self.unsettled.as_ref() {
            struct_ser.serialize_field("unsettled", v)?;
        }
        if let Some(v) = self.withdrawn.as_ref() {
            struct_ser.serialize_field("withdrawn", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PaymentState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "state",
            "rate",
            "balance",
            "unsettled",
            "withdrawn",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            State,
            Rate,
            Balance,
            Unsettled,
            Withdrawn,
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
                            "state" => Ok(GeneratedField::State),
                            "rate" => Ok(GeneratedField::Rate),
                            "balance" => Ok(GeneratedField::Balance),
                            "unsettled" => Ok(GeneratedField::Unsettled),
                            "withdrawn" => Ok(GeneratedField::Withdrawn),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PaymentState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.types.v1.PaymentState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PaymentState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut state__ = None;
                let mut rate__ = None;
                let mut balance__ = None;
                let mut unsettled__ = None;
                let mut withdrawn__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<State>()? as i32);
                        }
                        GeneratedField::Rate => {
                            if rate__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rate"));
                            }
                            rate__ = map_.next_value()?;
                        }
                        GeneratedField::Balance => {
                            if balance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("balance"));
                            }
                            balance__ = map_.next_value()?;
                        }
                        GeneratedField::Unsettled => {
                            if unsettled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unsettled"));
                            }
                            unsettled__ = map_.next_value()?;
                        }
                        GeneratedField::Withdrawn => {
                            if withdrawn__.is_some() {
                                return Err(serde::de::Error::duplicate_field("withdrawn"));
                            }
                            withdrawn__ = map_.next_value()?;
                        }
                    }
                }
                Ok(PaymentState {
                    owner: owner__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    rate: rate__,
                    balance: balance__,
                    unsettled: unsettled__,
                    withdrawn: withdrawn__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.types.v1.PaymentState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for State {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Invalid => "invalid",
            Self::Open => "open",
            Self::Closed => "closed",
            Self::Overdrawn => "overdrawn",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for State {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "invalid",
            "open",
            "closed",
            "overdrawn",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = State;

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
                    "invalid" => Ok(State::Invalid),
                    "open" => Ok(State::Open),
                    "closed" => Ok(State::Closed),
                    "overdrawn" => Ok(State::Overdrawn),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}

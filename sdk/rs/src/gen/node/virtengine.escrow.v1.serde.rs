// @generated
impl serde::Serialize for DepositAuthorization {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.spend_limit.is_some() {
            len += 1;
        }
        if !self.scopes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.DepositAuthorization", len)?;
        if let Some(v) = self.spend_limit.as_ref() {
            struct_ser.serialize_field("spendLimit", v)?;
        }
        if !self.scopes.is_empty() {
            let v = self.scopes.iter().cloned().map(|v| {
                deposit_authorization::Scope::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("scopes", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DepositAuthorization {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "spend_limit",
            "spendLimit",
            "scopes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SpendLimit,
            Scopes,
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
                            "spendLimit" | "spend_limit" => Ok(GeneratedField::SpendLimit),
                            "scopes" => Ok(GeneratedField::Scopes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DepositAuthorization;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.DepositAuthorization")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DepositAuthorization, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut spend_limit__ = None;
                let mut scopes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SpendLimit => {
                            if spend_limit__.is_some() {
                                return Err(serde::de::Error::duplicate_field("spendLimit"));
                            }
                            spend_limit__ = map_.next_value()?;
                        }
                        GeneratedField::Scopes => {
                            if scopes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopes"));
                            }
                            scopes__ = Some(map_.next_value::<Vec<deposit_authorization::Scope>>()?.into_iter().map(|x| x as i32).collect());
                        }
                    }
                }
                Ok(DepositAuthorization {
                    spend_limit: spend_limit__,
                    scopes: scopes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.DepositAuthorization", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for deposit_authorization::Scope {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Invalid => "invalid",
            Self::Deployment => "deployment",
            Self::Bid => "bid",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for deposit_authorization::Scope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "invalid",
            "deployment",
            "bid",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = deposit_authorization::Scope;

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
                    "invalid" => Ok(deposit_authorization::Scope::Invalid),
                    "deployment" => Ok(deposit_authorization::Scope::Deployment),
                    "bid" => Ok(deposit_authorization::Scope::Bid),
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
        if !self.accounts.is_empty() {
            len += 1;
        }
        if !self.payments.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.GenesisState", len)?;
        if !self.accounts.is_empty() {
            struct_ser.serialize_field("accounts", &self.accounts)?;
        }
        if !self.payments.is_empty() {
            struct_ser.serialize_field("payments", &self.payments)?;
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
            "accounts",
            "payments",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Accounts,
            Payments,
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
                            "accounts" => Ok(GeneratedField::Accounts),
                            "payments" => Ok(GeneratedField::Payments),
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
                formatter.write_str("struct virtengine.escrow.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut accounts__ = None;
                let mut payments__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Accounts => {
                            if accounts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accounts"));
                            }
                            accounts__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Payments => {
                            if payments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payments"));
                            }
                            payments__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    accounts: accounts__.unwrap_or_default(),
                    payments: payments__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAccountDeposit {
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
        if self.deposit.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.MsgAccountDeposit", len)?;
        if !self.signer.is_empty() {
            struct_ser.serialize_field("signer", &self.signer)?;
        }
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if let Some(v) = self.deposit.as_ref() {
            struct_ser.serialize_field("deposit", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAccountDeposit {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "signer",
            "id",
            "deposit",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Signer,
            Id,
            Deposit,
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
                            "deposit" => Ok(GeneratedField::Deposit),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAccountDeposit;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.MsgAccountDeposit")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAccountDeposit, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut signer__ = None;
                let mut id__ = None;
                let mut deposit__ = None;
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
                        GeneratedField::Deposit => {
                            if deposit__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deposit"));
                            }
                            deposit__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgAccountDeposit {
                    signer: signer__.unwrap_or_default(),
                    id: id__,
                    deposit: deposit__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.MsgAccountDeposit", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAccountDepositResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.escrow.v1.MsgAccountDepositResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAccountDepositResponse {
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
            type Value = MsgAccountDepositResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.MsgAccountDepositResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAccountDepositResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgAccountDepositResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.MsgAccountDepositResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAccountsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.state.is_empty() {
            len += 1;
        }
        if !self.xid.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.QueryAccountsRequest", len)?;
        if !self.state.is_empty() {
            struct_ser.serialize_field("state", &self.state)?;
        }
        if !self.xid.is_empty() {
            struct_ser.serialize_field("xid", &self.xid)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAccountsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "xid",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            Xid,
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
                            "xid" => Ok(GeneratedField::Xid),
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
            type Value = QueryAccountsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.QueryAccountsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAccountsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut xid__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Xid => {
                            if xid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("xid"));
                            }
                            xid__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAccountsRequest {
                    state: state__.unwrap_or_default(),
                    xid: xid__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.QueryAccountsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAccountsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.accounts.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.QueryAccountsResponse", len)?;
        if !self.accounts.is_empty() {
            struct_ser.serialize_field("accounts", &self.accounts)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAccountsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "accounts",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Accounts,
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
                            "accounts" => Ok(GeneratedField::Accounts),
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
            type Value = QueryAccountsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.QueryAccountsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAccountsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut accounts__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Accounts => {
                            if accounts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accounts"));
                            }
                            accounts__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAccountsResponse {
                    accounts: accounts__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.QueryAccountsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPaymentsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.state.is_empty() {
            len += 1;
        }
        if !self.xid.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.QueryPaymentsRequest", len)?;
        if !self.state.is_empty() {
            struct_ser.serialize_field("state", &self.state)?;
        }
        if !self.xid.is_empty() {
            struct_ser.serialize_field("xid", &self.xid)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPaymentsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "xid",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            Xid,
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
                            "xid" => Ok(GeneratedField::Xid),
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
            type Value = QueryPaymentsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.QueryPaymentsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPaymentsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut xid__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Xid => {
                            if xid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("xid"));
                            }
                            xid__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryPaymentsRequest {
                    state: state__.unwrap_or_default(),
                    xid: xid__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.QueryPaymentsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryPaymentsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.payments.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.escrow.v1.QueryPaymentsResponse", len)?;
        if !self.payments.is_empty() {
            struct_ser.serialize_field("payments", &self.payments)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryPaymentsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "payments",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Payments,
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
                            "payments" => Ok(GeneratedField::Payments),
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
            type Value = QueryPaymentsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.escrow.v1.QueryPaymentsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryPaymentsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut payments__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Payments => {
                            if payments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payments"));
                            }
                            payments__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryPaymentsResponse {
                    payments: payments__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.escrow.v1.QueryPaymentsResponse", FIELDS, GeneratedVisitor)
    }
}

// @generated
impl serde::Serialize for AccountState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "ACCOUNT_STATE_UNSPECIFIED",
            Self::Active => "ACCOUNT_STATE_ACTIVE",
            Self::Suspended => "ACCOUNT_STATE_SUSPENDED",
            Self::Terminated => "ACCOUNT_STATE_TERMINATED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for AccountState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ACCOUNT_STATE_UNSPECIFIED",
            "ACCOUNT_STATE_ACTIVE",
            "ACCOUNT_STATE_SUSPENDED",
            "ACCOUNT_STATE_TERMINATED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AccountState;

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
                    "ACCOUNT_STATE_UNSPECIFIED" => Ok(AccountState::Unspecified),
                    "ACCOUNT_STATE_ACTIVE" => Ok(AccountState::Active),
                    "ACCOUNT_STATE_SUSPENDED" => Ok(AccountState::Suspended),
                    "ACCOUNT_STATE_TERMINATED" => Ok(AccountState::Terminated),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for AccountStateRecord {
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
        if self.state != 0 {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.modified_by.is_empty() {
            len += 1;
        }
        if self.modified_at != 0 {
            len += 1;
        }
        if self.previous_state != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.AccountStateRecord", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if self.state != 0 {
            let v = AccountState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.modified_by.is_empty() {
            struct_ser.serialize_field("modifiedBy", &self.modified_by)?;
        }
        if self.modified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("modifiedAt", ToString::to_string(&self.modified_at).as_str())?;
        }
        if self.previous_state != 0 {
            let v = AccountState::try_from(self.previous_state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_state)))?;
            struct_ser.serialize_field("previousState", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AccountStateRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "state",
            "reason",
            "modified_by",
            "modifiedBy",
            "modified_at",
            "modifiedAt",
            "previous_state",
            "previousState",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            State,
            Reason,
            ModifiedBy,
            ModifiedAt,
            PreviousState,
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
                            "state" => Ok(GeneratedField::State),
                            "reason" => Ok(GeneratedField::Reason),
                            "modifiedBy" | "modified_by" => Ok(GeneratedField::ModifiedBy),
                            "modifiedAt" | "modified_at" => Ok(GeneratedField::ModifiedAt),
                            "previousState" | "previous_state" => Ok(GeneratedField::PreviousState),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AccountStateRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.AccountStateRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AccountStateRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut state__ = None;
                let mut reason__ = None;
                let mut modified_by__ = None;
                let mut modified_at__ = None;
                let mut previous_state__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<AccountState>()? as i32);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModifiedBy => {
                            if modified_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modifiedBy"));
                            }
                            modified_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModifiedAt => {
                            if modified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modifiedAt"));
                            }
                            modified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousState => {
                            if previous_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousState"));
                            }
                            previous_state__ = Some(map_.next_value::<AccountState>()? as i32);
                        }
                    }
                }
                Ok(AccountStateRecord {
                    address: address__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    modified_by: modified_by__.unwrap_or_default(),
                    modified_at: modified_at__.unwrap_or_default(),
                    previous_state: previous_state__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.AccountStateRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AccountStateStore {
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
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.modified_by.is_empty() {
            len += 1;
        }
        if self.modified_at != 0 {
            len += 1;
        }
        if self.previous_state != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.AccountStateStore", len)?;
        if self.state != 0 {
            struct_ser.serialize_field("state", &self.state)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.modified_by.is_empty() {
            struct_ser.serialize_field("modifiedBy", &self.modified_by)?;
        }
        if self.modified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("modifiedAt", ToString::to_string(&self.modified_at).as_str())?;
        }
        if self.previous_state != 0 {
            struct_ser.serialize_field("previousState", &self.previous_state)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AccountStateStore {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "reason",
            "modified_by",
            "modifiedBy",
            "modified_at",
            "modifiedAt",
            "previous_state",
            "previousState",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            Reason,
            ModifiedBy,
            ModifiedAt,
            PreviousState,
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
                            "reason" => Ok(GeneratedField::Reason),
                            "modifiedBy" | "modified_by" => Ok(GeneratedField::ModifiedBy),
                            "modifiedAt" | "modified_at" => Ok(GeneratedField::ModifiedAt),
                            "previousState" | "previous_state" => Ok(GeneratedField::PreviousState),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AccountStateStore;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.AccountStateStore")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AccountStateStore, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut reason__ = None;
                let mut modified_by__ = None;
                let mut modified_at__ = None;
                let mut previous_state__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModifiedBy => {
                            if modified_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modifiedBy"));
                            }
                            modified_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModifiedAt => {
                            if modified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modifiedAt"));
                            }
                            modified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousState => {
                            if previous_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousState"));
                            }
                            previous_state__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AccountStateStore {
                    state: state__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    modified_by: modified_by__.unwrap_or_default(),
                    modified_at: modified_at__.unwrap_or_default(),
                    previous_state: previous_state__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.AccountStateStore", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventAccountStateChanged {
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
        if !self.previous_state.is_empty() {
            len += 1;
        }
        if !self.new_state.is_empty() {
            len += 1;
        }
        if !self.modified_by.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.EventAccountStateChanged", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.previous_state.is_empty() {
            struct_ser.serialize_field("previousState", &self.previous_state)?;
        }
        if !self.new_state.is_empty() {
            struct_ser.serialize_field("newState", &self.new_state)?;
        }
        if !self.modified_by.is_empty() {
            struct_ser.serialize_field("modifiedBy", &self.modified_by)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventAccountStateChanged {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "previous_state",
            "previousState",
            "new_state",
            "newState",
            "modified_by",
            "modifiedBy",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            PreviousState,
            NewState,
            ModifiedBy,
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
                            "previousState" | "previous_state" => Ok(GeneratedField::PreviousState),
                            "newState" | "new_state" => Ok(GeneratedField::NewState),
                            "modifiedBy" | "modified_by" => Ok(GeneratedField::ModifiedBy),
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
            type Value = EventAccountStateChanged;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.EventAccountStateChanged")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventAccountStateChanged, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut previous_state__ = None;
                let mut new_state__ = None;
                let mut modified_by__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PreviousState => {
                            if previous_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousState"));
                            }
                            previous_state__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewState => {
                            if new_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newState"));
                            }
                            new_state__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModifiedBy => {
                            if modified_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modifiedBy"));
                            }
                            modified_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventAccountStateChanged {
                    address: address__.unwrap_or_default(),
                    previous_state: previous_state__.unwrap_or_default(),
                    new_state: new_state__.unwrap_or_default(),
                    modified_by: modified_by__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.EventAccountStateChanged", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventAdminNominated {
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
        if !self.nominated_by.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.EventAdminNominated", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.nominated_by.is_empty() {
            struct_ser.serialize_field("nominatedBy", &self.nominated_by)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventAdminNominated {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "nominated_by",
            "nominatedBy",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            NominatedBy,
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
                            "nominatedBy" | "nominated_by" => Ok(GeneratedField::NominatedBy),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventAdminNominated;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.EventAdminNominated")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventAdminNominated, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut nominated_by__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NominatedBy => {
                            if nominated_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nominatedBy"));
                            }
                            nominated_by__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventAdminNominated {
                    address: address__.unwrap_or_default(),
                    nominated_by: nominated_by__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.EventAdminNominated", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventRoleAssigned {
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
        if !self.role.is_empty() {
            len += 1;
        }
        if !self.assigned_by.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.EventRoleAssigned", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        if !self.assigned_by.is_empty() {
            struct_ser.serialize_field("assignedBy", &self.assigned_by)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventRoleAssigned {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "role",
            "assigned_by",
            "assignedBy",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Role,
            AssignedBy,
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
                            "role" => Ok(GeneratedField::Role),
                            "assignedBy" | "assigned_by" => Ok(GeneratedField::AssignedBy),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventRoleAssigned;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.EventRoleAssigned")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventRoleAssigned, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut role__ = None;
                let mut assigned_by__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AssignedBy => {
                            if assigned_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedBy"));
                            }
                            assigned_by__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventRoleAssigned {
                    address: address__.unwrap_or_default(),
                    role: role__.unwrap_or_default(),
                    assigned_by: assigned_by__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.EventRoleAssigned", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventRoleRevoked {
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
        if !self.role.is_empty() {
            len += 1;
        }
        if !self.revoked_by.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.EventRoleRevoked", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        if !self.revoked_by.is_empty() {
            struct_ser.serialize_field("revokedBy", &self.revoked_by)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventRoleRevoked {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "role",
            "revoked_by",
            "revokedBy",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Role,
            RevokedBy,
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
                            "role" => Ok(GeneratedField::Role),
                            "revokedBy" | "revoked_by" => Ok(GeneratedField::RevokedBy),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventRoleRevoked;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.EventRoleRevoked")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventRoleRevoked, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut role__ = None;
                let mut revoked_by__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedBy => {
                            if revoked_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedBy"));
                            }
                            revoked_by__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventRoleRevoked {
                    address: address__.unwrap_or_default(),
                    role: role__.unwrap_or_default(),
                    revoked_by: revoked_by__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.EventRoleRevoked", FIELDS, GeneratedVisitor)
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
        if !self.genesis_accounts.is_empty() {
            len += 1;
        }
        if !self.role_assignments.is_empty() {
            len += 1;
        }
        if !self.account_states.is_empty() {
            len += 1;
        }
        if self.params.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.GenesisState", len)?;
        if !self.genesis_accounts.is_empty() {
            struct_ser.serialize_field("genesisAccounts", &self.genesis_accounts)?;
        }
        if !self.role_assignments.is_empty() {
            struct_ser.serialize_field("roleAssignments", &self.role_assignments)?;
        }
        if !self.account_states.is_empty() {
            struct_ser.serialize_field("accountStates", &self.account_states)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
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
            "genesis_accounts",
            "genesisAccounts",
            "role_assignments",
            "roleAssignments",
            "account_states",
            "accountStates",
            "params",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            GenesisAccounts,
            RoleAssignments,
            AccountStates,
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
                            "genesisAccounts" | "genesis_accounts" => Ok(GeneratedField::GenesisAccounts),
                            "roleAssignments" | "role_assignments" => Ok(GeneratedField::RoleAssignments),
                            "accountStates" | "account_states" => Ok(GeneratedField::AccountStates),
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
            type Value = GenesisState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut genesis_accounts__ = None;
                let mut role_assignments__ = None;
                let mut account_states__ = None;
                let mut params__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::GenesisAccounts => {
                            if genesis_accounts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("genesisAccounts"));
                            }
                            genesis_accounts__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RoleAssignments => {
                            if role_assignments__.is_some() {
                                return Err(serde::de::Error::duplicate_field("roleAssignments"));
                            }
                            role_assignments__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountStates => {
                            if account_states__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountStates"));
                            }
                            account_states__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                    }
                }
                Ok(GenesisState {
                    genesis_accounts: genesis_accounts__.unwrap_or_default(),
                    role_assignments: role_assignments__.unwrap_or_default(),
                    account_states: account_states__.unwrap_or_default(),
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAssignRole {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.address.is_empty() {
            len += 1;
        }
        if !self.role.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgAssignRole", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAssignRole {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "address",
            "role",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            Address,
            Role,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "address" => Ok(GeneratedField::Address),
                            "role" => Ok(GeneratedField::Role),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAssignRole;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgAssignRole")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAssignRole, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut address__ = None;
                let mut role__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgAssignRole {
                    sender: sender__.unwrap_or_default(),
                    address: address__.unwrap_or_default(),
                    role: role__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgAssignRole", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAssignRoleResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgAssignRoleResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAssignRoleResponse {
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
            type Value = MsgAssignRoleResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgAssignRoleResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAssignRoleResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgAssignRoleResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgAssignRoleResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgNominateAdmin {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgNominateAdmin", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgNominateAdmin {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "address",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            Address,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "address" => Ok(GeneratedField::Address),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgNominateAdmin;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgNominateAdmin")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgNominateAdmin, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgNominateAdmin {
                    sender: sender__.unwrap_or_default(),
                    address: address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgNominateAdmin", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgNominateAdminResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgNominateAdminResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgNominateAdminResponse {
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
            type Value = MsgNominateAdminResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgNominateAdminResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgNominateAdminResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgNominateAdminResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgNominateAdminResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeRole {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.address.is_empty() {
            len += 1;
        }
        if !self.role.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgRevokeRole", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeRole {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "address",
            "role",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            Address,
            Role,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "address" => Ok(GeneratedField::Address),
                            "role" => Ok(GeneratedField::Role),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeRole;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgRevokeRole")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeRole, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut address__ = None;
                let mut role__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRevokeRole {
                    sender: sender__.unwrap_or_default(),
                    address: address__.unwrap_or_default(),
                    role: role__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgRevokeRole", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeRoleResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgRevokeRoleResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeRoleResponse {
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
            type Value = MsgRevokeRoleResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgRevokeRoleResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeRoleResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRevokeRoleResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgRevokeRoleResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSetAccountState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.sender.is_empty() {
            len += 1;
        }
        if !self.address.is_empty() {
            len += 1;
        }
        if !self.state.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if self.mfa_proof.is_some() {
            len += 1;
        }
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgSetAccountState", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.state.is_empty() {
            struct_ser.serialize_field("state", &self.state)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if let Some(v) = self.mfa_proof.as_ref() {
            struct_ser.serialize_field("mfaProof", v)?;
        }
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSetAccountState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "address",
            "state",
            "reason",
            "mfa_proof",
            "mfaProof",
            "device_fingerprint",
            "deviceFingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            Address,
            State,
            Reason,
            MfaProof,
            DeviceFingerprint,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "address" => Ok(GeneratedField::Address),
                            "state" => Ok(GeneratedField::State),
                            "reason" => Ok(GeneratedField::Reason),
                            "mfaProof" | "mfa_proof" => Ok(GeneratedField::MfaProof),
                            "deviceFingerprint" | "device_fingerprint" => Ok(GeneratedField::DeviceFingerprint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSetAccountState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgSetAccountState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSetAccountState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut address__ = None;
                let mut state__ = None;
                let mut reason__ = None;
                let mut mfa_proof__ = None;
                let mut device_fingerprint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MfaProof => {
                            if mfa_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaProof"));
                            }
                            mfa_proof__ = map_.next_value()?;
                        }
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSetAccountState {
                    sender: sender__.unwrap_or_default(),
                    address: address__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    mfa_proof: mfa_proof__,
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgSetAccountState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSetAccountStateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgSetAccountStateResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSetAccountStateResponse {
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
            type Value = MsgSetAccountStateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.MsgSetAccountStateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSetAccountStateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgSetAccountStateResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.MsgSetAccountStateResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.roles.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.roles.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.roles.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.roles.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.roles.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if self.max_roles_per_account != 0 {
            len += 1;
        }
        if self.allow_self_revoke {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.Params", len)?;
        if self.max_roles_per_account != 0 {
            struct_ser.serialize_field("maxRolesPerAccount", &self.max_roles_per_account)?;
        }
        if self.allow_self_revoke {
            struct_ser.serialize_field("allowSelfRevoke", &self.allow_self_revoke)?;
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
            "max_roles_per_account",
            "maxRolesPerAccount",
            "allow_self_revoke",
            "allowSelfRevoke",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MaxRolesPerAccount,
            AllowSelfRevoke,
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
                            "maxRolesPerAccount" | "max_roles_per_account" => Ok(GeneratedField::MaxRolesPerAccount),
                            "allowSelfRevoke" | "allow_self_revoke" => Ok(GeneratedField::AllowSelfRevoke),
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
                formatter.write_str("struct virtengine.roles.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut max_roles_per_account__ = None;
                let mut allow_self_revoke__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MaxRolesPerAccount => {
                            if max_roles_per_account__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRolesPerAccount"));
                            }
                            max_roles_per_account__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AllowSelfRevoke => {
                            if allow_self_revoke__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowSelfRevoke"));
                            }
                            allow_self_revoke__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Params {
                    max_roles_per_account: max_roles_per_account__.unwrap_or_default(),
                    allow_self_revoke: allow_self_revoke__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ParamsStore {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.max_roles_per_account != 0 {
            len += 1;
        }
        if self.allow_self_revoke {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.ParamsStore", len)?;
        if self.max_roles_per_account != 0 {
            struct_ser.serialize_field("maxRolesPerAccount", &self.max_roles_per_account)?;
        }
        if self.allow_self_revoke {
            struct_ser.serialize_field("allowSelfRevoke", &self.allow_self_revoke)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ParamsStore {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "max_roles_per_account",
            "maxRolesPerAccount",
            "allow_self_revoke",
            "allowSelfRevoke",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MaxRolesPerAccount,
            AllowSelfRevoke,
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
                            "maxRolesPerAccount" | "max_roles_per_account" => Ok(GeneratedField::MaxRolesPerAccount),
                            "allowSelfRevoke" | "allow_self_revoke" => Ok(GeneratedField::AllowSelfRevoke),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ParamsStore;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.ParamsStore")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ParamsStore, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut max_roles_per_account__ = None;
                let mut allow_self_revoke__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MaxRolesPerAccount => {
                            if max_roles_per_account__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRolesPerAccount"));
                            }
                            max_roles_per_account__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AllowSelfRevoke => {
                            if allow_self_revoke__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowSelfRevoke"));
                            }
                            allow_self_revoke__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ParamsStore {
                    max_roles_per_account: max_roles_per_account__.unwrap_or_default(),
                    allow_self_revoke: allow_self_revoke__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.ParamsStore", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAccountRolesRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryAccountRolesRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAccountRolesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
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
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAccountRolesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryAccountRolesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAccountRolesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAccountRolesRequest {
                    address: address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryAccountRolesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAccountRolesResponse {
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
        if !self.roles.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryAccountRolesResponse", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.roles.is_empty() {
            struct_ser.serialize_field("roles", &self.roles)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAccountRolesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "roles",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Roles,
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
                            "roles" => Ok(GeneratedField::Roles),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAccountRolesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryAccountRolesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAccountRolesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut roles__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Roles => {
                            if roles__.is_some() {
                                return Err(serde::de::Error::duplicate_field("roles"));
                            }
                            roles__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAccountRolesResponse {
                    address: address__.unwrap_or_default(),
                    roles: roles__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryAccountRolesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAccountStateRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryAccountStateRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAccountStateRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
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
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAccountStateRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryAccountStateRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAccountStateRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAccountStateRequest {
                    address: address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryAccountStateRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAccountStateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.account_state.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryAccountStateResponse", len)?;
        if let Some(v) = self.account_state.as_ref() {
            struct_ser.serialize_field("accountState", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAccountStateResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_state",
            "accountState",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountState,
            Found,
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
                            "accountState" | "account_state" => Ok(GeneratedField::AccountState),
                            "found" => Ok(GeneratedField::Found),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAccountStateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryAccountStateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAccountStateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_state__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountState => {
                            if account_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountState"));
                            }
                            account_state__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAccountStateResponse {
                    account_state: account_state__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryAccountStateResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryGenesisAccountsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryGenesisAccountsRequest", len)?;
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryGenesisAccountsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = QueryGenesisAccountsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryGenesisAccountsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryGenesisAccountsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryGenesisAccountsRequest {
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryGenesisAccountsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryGenesisAccountsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.addresses.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryGenesisAccountsResponse", len)?;
        if !self.addresses.is_empty() {
            struct_ser.serialize_field("addresses", &self.addresses)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryGenesisAccountsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "addresses",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Addresses,
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
                            "addresses" => Ok(GeneratedField::Addresses),
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
            type Value = QueryGenesisAccountsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryGenesisAccountsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryGenesisAccountsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut addresses__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Addresses => {
                            if addresses__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addresses"));
                            }
                            addresses__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryGenesisAccountsResponse {
                    addresses: addresses__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryGenesisAccountsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryHasRoleRequest {
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
        if !self.role.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryHasRoleRequest", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryHasRoleRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "role",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Role,
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
                            "role" => Ok(GeneratedField::Role),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryHasRoleRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryHasRoleRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryHasRoleRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut role__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryHasRoleRequest {
                    address: address__.unwrap_or_default(),
                    role: role__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryHasRoleRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryHasRoleResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.has_role {
            len += 1;
        }
        if self.assignment.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryHasRoleResponse", len)?;
        if self.has_role {
            struct_ser.serialize_field("hasRole", &self.has_role)?;
        }
        if let Some(v) = self.assignment.as_ref() {
            struct_ser.serialize_field("assignment", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryHasRoleResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "has_role",
            "hasRole",
            "assignment",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            HasRole,
            Assignment,
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
                            "hasRole" | "has_role" => Ok(GeneratedField::HasRole),
                            "assignment" => Ok(GeneratedField::Assignment),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryHasRoleResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryHasRoleResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryHasRoleResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut has_role__ = None;
                let mut assignment__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::HasRole => {
                            if has_role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hasRole"));
                            }
                            has_role__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Assignment => {
                            if assignment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignment"));
                            }
                            assignment__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryHasRoleResponse {
                    has_role: has_role__.unwrap_or_default(),
                    assignment: assignment__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryHasRoleResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.roles.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.roles.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.roles.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.roles.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRoleMembersRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.role.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryRoleMembersRequest", len)?;
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRoleMembersRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "role",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Role,
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
                            "role" => Ok(GeneratedField::Role),
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
            type Value = QueryRoleMembersRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryRoleMembersRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRoleMembersRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut role__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryRoleMembersRequest {
                    role: role__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryRoleMembersRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRoleMembersResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.role.is_empty() {
            len += 1;
        }
        if !self.members.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.QueryRoleMembersResponse", len)?;
        if !self.role.is_empty() {
            struct_ser.serialize_field("role", &self.role)?;
        }
        if !self.members.is_empty() {
            struct_ser.serialize_field("members", &self.members)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRoleMembersResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "role",
            "members",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Role,
            Members,
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
                            "role" => Ok(GeneratedField::Role),
                            "members" => Ok(GeneratedField::Members),
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
            type Value = QueryRoleMembersResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.QueryRoleMembersResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRoleMembersResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut role__ = None;
                let mut members__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Members => {
                            if members__.is_some() {
                                return Err(serde::de::Error::duplicate_field("members"));
                            }
                            members__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryRoleMembersResponse {
                    role: role__.unwrap_or_default(),
                    members: members__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.QueryRoleMembersResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Role {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "ROLE_UNSPECIFIED",
            Self::GenesisAccount => "ROLE_GENESIS_ACCOUNT",
            Self::Administrator => "ROLE_ADMINISTRATOR",
            Self::Moderator => "ROLE_MODERATOR",
            Self::Validator => "ROLE_VALIDATOR",
            Self::ServiceProvider => "ROLE_SERVICE_PROVIDER",
            Self::Customer => "ROLE_CUSTOMER",
            Self::SupportAgent => "ROLE_SUPPORT_AGENT",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for Role {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ROLE_UNSPECIFIED",
            "ROLE_GENESIS_ACCOUNT",
            "ROLE_ADMINISTRATOR",
            "ROLE_MODERATOR",
            "ROLE_VALIDATOR",
            "ROLE_SERVICE_PROVIDER",
            "ROLE_CUSTOMER",
            "ROLE_SUPPORT_AGENT",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Role;

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
                    "ROLE_UNSPECIFIED" => Ok(Role::Unspecified),
                    "ROLE_GENESIS_ACCOUNT" => Ok(Role::GenesisAccount),
                    "ROLE_ADMINISTRATOR" => Ok(Role::Administrator),
                    "ROLE_MODERATOR" => Ok(Role::Moderator),
                    "ROLE_VALIDATOR" => Ok(Role::Validator),
                    "ROLE_SERVICE_PROVIDER" => Ok(Role::ServiceProvider),
                    "ROLE_CUSTOMER" => Ok(Role::Customer),
                    "ROLE_SUPPORT_AGENT" => Ok(Role::SupportAgent),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for RoleAssignment {
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
        if self.role != 0 {
            len += 1;
        }
        if !self.assigned_by.is_empty() {
            len += 1;
        }
        if self.assigned_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.RoleAssignment", len)?;
        if !self.address.is_empty() {
            struct_ser.serialize_field("address", &self.address)?;
        }
        if self.role != 0 {
            let v = Role::try_from(self.role)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.role)))?;
            struct_ser.serialize_field("role", &v)?;
        }
        if !self.assigned_by.is_empty() {
            struct_ser.serialize_field("assignedBy", &self.assigned_by)?;
        }
        if self.assigned_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("assignedAt", ToString::to_string(&self.assigned_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RoleAssignment {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "address",
            "role",
            "assigned_by",
            "assignedBy",
            "assigned_at",
            "assignedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Address,
            Role,
            AssignedBy,
            AssignedAt,
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
                            "role" => Ok(GeneratedField::Role),
                            "assignedBy" | "assigned_by" => Ok(GeneratedField::AssignedBy),
                            "assignedAt" | "assigned_at" => Ok(GeneratedField::AssignedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RoleAssignment;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.RoleAssignment")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RoleAssignment, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut address__ = None;
                let mut role__ = None;
                let mut assigned_by__ = None;
                let mut assigned_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Address => {
                            if address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("address"));
                            }
                            address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Role => {
                            if role__.is_some() {
                                return Err(serde::de::Error::duplicate_field("role"));
                            }
                            role__ = Some(map_.next_value::<Role>()? as i32);
                        }
                        GeneratedField::AssignedBy => {
                            if assigned_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedBy"));
                            }
                            assigned_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AssignedAt => {
                            if assigned_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedAt"));
                            }
                            assigned_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(RoleAssignment {
                    address: address__.unwrap_or_default(),
                    role: role__.unwrap_or_default(),
                    assigned_by: assigned_by__.unwrap_or_default(),
                    assigned_at: assigned_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.RoleAssignment", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RoleAssignmentStore {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.assigned_by.is_empty() {
            len += 1;
        }
        if self.assigned_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.roles.v1.RoleAssignmentStore", len)?;
        if !self.assigned_by.is_empty() {
            struct_ser.serialize_field("assignedBy", &self.assigned_by)?;
        }
        if self.assigned_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("assignedAt", ToString::to_string(&self.assigned_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RoleAssignmentStore {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "assigned_by",
            "assignedBy",
            "assigned_at",
            "assignedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AssignedBy,
            AssignedAt,
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
                            "assignedBy" | "assigned_by" => Ok(GeneratedField::AssignedBy),
                            "assignedAt" | "assigned_at" => Ok(GeneratedField::AssignedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RoleAssignmentStore;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.roles.v1.RoleAssignmentStore")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RoleAssignmentStore, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut assigned_by__ = None;
                let mut assigned_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AssignedBy => {
                            if assigned_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedBy"));
                            }
                            assigned_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AssignedAt => {
                            if assigned_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("assignedAt"));
                            }
                            assigned_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(RoleAssignmentStore {
                    assigned_by: assigned_by__.unwrap_or_default(),
                    assigned_at: assigned_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.roles.v1.RoleAssignmentStore", FIELDS, GeneratedVisitor)
    }
}

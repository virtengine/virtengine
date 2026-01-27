// @generated
impl serde::Serialize for AttributesFilters {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.auditors.is_empty() {
            len += 1;
        }
        if !self.owners.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.AttributesFilters", len)?;
        if !self.auditors.is_empty() {
            struct_ser.serialize_field("auditors", &self.auditors)?;
        }
        if !self.owners.is_empty() {
            struct_ser.serialize_field("owners", &self.owners)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AttributesFilters {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "auditors",
            "owners",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Auditors,
            Owners,
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
                            "auditors" => Ok(GeneratedField::Auditors),
                            "owners" => Ok(GeneratedField::Owners),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AttributesFilters;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.AttributesFilters")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AttributesFilters, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut auditors__ = None;
                let mut owners__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Auditors => {
                            if auditors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditors"));
                            }
                            auditors__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Owners => {
                            if owners__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owners"));
                            }
                            owners__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(AttributesFilters {
                    auditors: auditors__.unwrap_or_default(),
                    owners: owners__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.AttributesFilters", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AuditedAttributesStore {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.AuditedAttributesStore", len)?;
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AuditedAttributesStore {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Attributes,
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
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AuditedAttributesStore;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.AuditedAttributesStore")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AuditedAttributesStore, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(AuditedAttributesStore {
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.AuditedAttributesStore", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AuditedProvider {
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
        if !self.auditor.is_empty() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.AuditedProvider", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AuditedProvider {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "auditor",
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            Auditor,
            Attributes,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AuditedProvider;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.AuditedProvider")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AuditedProvider, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut auditor__ = None;
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(AuditedProvider {
                    owner: owner__.unwrap_or_default(),
                    auditor: auditor__.unwrap_or_default(),
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.AuditedProvider", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventTrustedAuditorCreated {
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
        if !self.auditor.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.EventTrustedAuditorCreated", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventTrustedAuditorCreated {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "auditor",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            Auditor,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventTrustedAuditorCreated;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.EventTrustedAuditorCreated")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventTrustedAuditorCreated, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut auditor__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventTrustedAuditorCreated {
                    owner: owner__.unwrap_or_default(),
                    auditor: auditor__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.EventTrustedAuditorCreated", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EventTrustedAuditorDeleted {
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
        if !self.auditor.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.EventTrustedAuditorDeleted", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EventTrustedAuditorDeleted {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "auditor",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            Auditor,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EventTrustedAuditorDeleted;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.EventTrustedAuditorDeleted")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EventTrustedAuditorDeleted, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut auditor__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EventTrustedAuditorDeleted {
                    owner: owner__.unwrap_or_default(),
                    auditor: auditor__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.EventTrustedAuditorDeleted", FIELDS, GeneratedVisitor)
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
        if !self.providers.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.GenesisState", len)?;
        if !self.providers.is_empty() {
            struct_ser.serialize_field("providers", &self.providers)?;
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
            "providers",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Providers,
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
                            "providers" => Ok(GeneratedField::Providers),
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
                formatter.write_str("struct virtengine.audit.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut providers__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Providers => {
                            if providers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providers"));
                            }
                            providers__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    providers: providers__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeleteProviderAttributes {
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
        if !self.auditor.is_empty() {
            len += 1;
        }
        if !self.keys.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.MsgDeleteProviderAttributes", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        if !self.keys.is_empty() {
            struct_ser.serialize_field("keys", &self.keys)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeleteProviderAttributes {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "auditor",
            "keys",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            Auditor,
            Keys,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            "keys" => Ok(GeneratedField::Keys),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgDeleteProviderAttributes;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.MsgDeleteProviderAttributes")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeleteProviderAttributes, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut auditor__ = None;
                let mut keys__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Keys => {
                            if keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("keys"));
                            }
                            keys__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDeleteProviderAttributes {
                    owner: owner__.unwrap_or_default(),
                    auditor: auditor__.unwrap_or_default(),
                    keys: keys__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.MsgDeleteProviderAttributes", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeleteProviderAttributesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.audit.v1.MsgDeleteProviderAttributesResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeleteProviderAttributesResponse {
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
            type Value = MsgDeleteProviderAttributesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.MsgDeleteProviderAttributesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeleteProviderAttributesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDeleteProviderAttributesResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.MsgDeleteProviderAttributesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSignProviderAttributes {
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
        if !self.auditor.is_empty() {
            len += 1;
        }
        if !self.attributes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.MsgSignProviderAttributes", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        if !self.attributes.is_empty() {
            struct_ser.serialize_field("attributes", &self.attributes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSignProviderAttributes {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "auditor",
            "attributes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
            Auditor,
            Attributes,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            "attributes" => Ok(GeneratedField::Attributes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSignProviderAttributes;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.MsgSignProviderAttributes")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSignProviderAttributes, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut auditor__ = None;
                let mut attributes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Attributes => {
                            if attributes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attributes"));
                            }
                            attributes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSignProviderAttributes {
                    owner: owner__.unwrap_or_default(),
                    auditor: auditor__.unwrap_or_default(),
                    attributes: attributes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.MsgSignProviderAttributes", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSignProviderAttributesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.audit.v1.MsgSignProviderAttributesResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSignProviderAttributesResponse {
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
            type Value = MsgSignProviderAttributesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.MsgSignProviderAttributesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSignProviderAttributesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgSignProviderAttributesResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.MsgSignProviderAttributesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAllProvidersAttributesRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.QueryAllProvidersAttributesRequest", len)?;
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAllProvidersAttributesRequest {
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
            type Value = QueryAllProvidersAttributesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.QueryAllProvidersAttributesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAllProvidersAttributesRequest, V::Error>
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
                Ok(QueryAllProvidersAttributesRequest {
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.QueryAllProvidersAttributesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAuditorAttributesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.auditor.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.QueryAuditorAttributesRequest", len)?;
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAuditorAttributesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "auditor",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Auditor,
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
                            "auditor" => Ok(GeneratedField::Auditor),
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
            type Value = QueryAuditorAttributesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.QueryAuditorAttributesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAuditorAttributesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut auditor__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAuditorAttributesRequest {
                    auditor: auditor__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.QueryAuditorAttributesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryProviderAttributesRequest {
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
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.QueryProviderAttributesRequest", len)?;
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryProviderAttributesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "owner",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Owner,
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
                            "owner" => Ok(GeneratedField::Owner),
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
            type Value = QueryProviderAttributesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.QueryProviderAttributesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryProviderAttributesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut owner__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryProviderAttributesRequest {
                    owner: owner__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.QueryProviderAttributesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryProviderAuditorRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.auditor.is_empty() {
            len += 1;
        }
        if !self.owner.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.QueryProviderAuditorRequest", len)?;
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryProviderAuditorRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "auditor",
            "owner",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Auditor,
            Owner,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            "owner" => Ok(GeneratedField::Owner),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryProviderAuditorRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.QueryProviderAuditorRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryProviderAuditorRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut auditor__ = None;
                let mut owner__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryProviderAuditorRequest {
                    auditor: auditor__.unwrap_or_default(),
                    owner: owner__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.QueryProviderAuditorRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryProviderRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.auditor.is_empty() {
            len += 1;
        }
        if !self.owner.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.QueryProviderRequest", len)?;
        if !self.auditor.is_empty() {
            struct_ser.serialize_field("auditor", &self.auditor)?;
        }
        if !self.owner.is_empty() {
            struct_ser.serialize_field("owner", &self.owner)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryProviderRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "auditor",
            "owner",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Auditor,
            Owner,
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
                            "auditor" => Ok(GeneratedField::Auditor),
                            "owner" => Ok(GeneratedField::Owner),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryProviderRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.QueryProviderRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryProviderRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut auditor__ = None;
                let mut owner__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Auditor => {
                            if auditor__.is_some() {
                                return Err(serde::de::Error::duplicate_field("auditor"));
                            }
                            auditor__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Owner => {
                            if owner__.is_some() {
                                return Err(serde::de::Error::duplicate_field("owner"));
                            }
                            owner__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryProviderRequest {
                    auditor: auditor__.unwrap_or_default(),
                    owner: owner__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.QueryProviderRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryProvidersResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.providers.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.audit.v1.QueryProvidersResponse", len)?;
        if !self.providers.is_empty() {
            struct_ser.serialize_field("providers", &self.providers)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryProvidersResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "providers",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Providers,
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
                            "providers" => Ok(GeneratedField::Providers),
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
            type Value = QueryProvidersResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.audit.v1.QueryProvidersResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryProvidersResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut providers__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Providers => {
                            if providers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providers"));
                            }
                            providers__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryProvidersResponse {
                    providers: providers__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.audit.v1.QueryProvidersResponse", FIELDS, GeneratedVisitor)
    }
}

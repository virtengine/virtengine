// @generated
impl serde::Serialize for Allocation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.allocation_id.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if !self.offering_id.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.customer_address.is_empty() {
            len += 1;
        }
        if self.state != 0 {
            len += 1;
        }
        if self.accepted_price != 0 {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.terminated_at.is_some() {
            len += 1;
        }
        if !self.state_reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.Allocation", len)?;
        if !self.allocation_id.is_empty() {
            struct_ser.serialize_field("allocationId", &self.allocation_id)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.customer_address.is_empty() {
            struct_ser.serialize_field("customerAddress", &self.customer_address)?;
        }
        if self.state != 0 {
            let v = AllocationState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if self.accepted_price != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("acceptedPrice", ToString::to_string(&self.accepted_price).as_str())?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if let Some(v) = self.terminated_at.as_ref() {
            struct_ser.serialize_field("terminatedAt", v)?;
        }
        if !self.state_reason.is_empty() {
            struct_ser.serialize_field("stateReason", &self.state_reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Allocation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "allocation_id",
            "allocationId",
            "order_id",
            "orderId",
            "offering_id",
            "offeringId",
            "provider_address",
            "providerAddress",
            "customer_address",
            "customerAddress",
            "state",
            "accepted_price",
            "acceptedPrice",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "terminated_at",
            "terminatedAt",
            "state_reason",
            "stateReason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AllocationId,
            OrderId,
            OfferingId,
            ProviderAddress,
            CustomerAddress,
            State,
            AcceptedPrice,
            CreatedAt,
            UpdatedAt,
            TerminatedAt,
            StateReason,
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
                            "allocationId" | "allocation_id" => Ok(GeneratedField::AllocationId),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "customerAddress" | "customer_address" => Ok(GeneratedField::CustomerAddress),
                            "state" => Ok(GeneratedField::State),
                            "acceptedPrice" | "accepted_price" => Ok(GeneratedField::AcceptedPrice),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "terminatedAt" | "terminated_at" => Ok(GeneratedField::TerminatedAt),
                            "stateReason" | "state_reason" => Ok(GeneratedField::StateReason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Allocation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.Allocation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Allocation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut allocation_id__ = None;
                let mut order_id__ = None;
                let mut offering_id__ = None;
                let mut provider_address__ = None;
                let mut customer_address__ = None;
                let mut state__ = None;
                let mut accepted_price__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut terminated_at__ = None;
                let mut state_reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AllocationId => {
                            if allocation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocationId"));
                            }
                            allocation_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
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
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map_.next_value::<AllocationState>()? as i32);
                        }
                        GeneratedField::AcceptedPrice => {
                            if accepted_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("acceptedPrice"));
                            }
                            accepted_price__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
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
                        GeneratedField::TerminatedAt => {
                            if terminated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("terminatedAt"));
                            }
                            terminated_at__ = map_.next_value()?;
                        }
                        GeneratedField::StateReason => {
                            if state_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("stateReason"));
                            }
                            state_reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Allocation {
                    allocation_id: allocation_id__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    offering_id: offering_id__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    customer_address: customer_address__.unwrap_or_default(),
                    state: state__.unwrap_or_default(),
                    accepted_price: accepted_price__.unwrap_or_default(),
                    created_at: created_at__,
                    updated_at: updated_at__,
                    terminated_at: terminated_at__,
                    state_reason: state_reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.Allocation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AllocationState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "ALLOCATION_STATE_UNSPECIFIED",
            Self::Pending => "ALLOCATION_STATE_PENDING",
            Self::Accepted => "ALLOCATION_STATE_ACCEPTED",
            Self::Provisioning => "ALLOCATION_STATE_PROVISIONING",
            Self::Active => "ALLOCATION_STATE_ACTIVE",
            Self::Suspended => "ALLOCATION_STATE_SUSPENDED",
            Self::Terminating => "ALLOCATION_STATE_TERMINATING",
            Self::Terminated => "ALLOCATION_STATE_TERMINATED",
            Self::Rejected => "ALLOCATION_STATE_REJECTED",
            Self::Failed => "ALLOCATION_STATE_FAILED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for AllocationState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ALLOCATION_STATE_UNSPECIFIED",
            "ALLOCATION_STATE_PENDING",
            "ALLOCATION_STATE_ACCEPTED",
            "ALLOCATION_STATE_PROVISIONING",
            "ALLOCATION_STATE_ACTIVE",
            "ALLOCATION_STATE_SUSPENDED",
            "ALLOCATION_STATE_TERMINATING",
            "ALLOCATION_STATE_TERMINATED",
            "ALLOCATION_STATE_REJECTED",
            "ALLOCATION_STATE_FAILED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AllocationState;

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
                    "ALLOCATION_STATE_UNSPECIFIED" => Ok(AllocationState::Unspecified),
                    "ALLOCATION_STATE_PENDING" => Ok(AllocationState::Pending),
                    "ALLOCATION_STATE_ACCEPTED" => Ok(AllocationState::Accepted),
                    "ALLOCATION_STATE_PROVISIONING" => Ok(AllocationState::Provisioning),
                    "ALLOCATION_STATE_ACTIVE" => Ok(AllocationState::Active),
                    "ALLOCATION_STATE_SUSPENDED" => Ok(AllocationState::Suspended),
                    "ALLOCATION_STATE_TERMINATING" => Ok(AllocationState::Terminating),
                    "ALLOCATION_STATE_TERMINATED" => Ok(AllocationState::Terminated),
                    "ALLOCATION_STATE_REJECTED" => Ok(AllocationState::Rejected),
                    "ALLOCATION_STATE_FAILED" => Ok(AllocationState::Failed),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for EncryptedProviderSecrets {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.envelope.is_some() {
            len += 1;
        }
        if !self.envelope_ref.is_empty() {
            len += 1;
        }
        if !self.recipient_key_ids.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.EncryptedProviderSecrets", len)?;
        if let Some(v) = self.envelope.as_ref() {
            struct_ser.serialize_field("envelope", v)?;
        }
        if !self.envelope_ref.is_empty() {
            struct_ser.serialize_field("envelopeRef", &self.envelope_ref)?;
        }
        if !self.recipient_key_ids.is_empty() {
            struct_ser.serialize_field("recipientKeyIds", &self.recipient_key_ids)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EncryptedProviderSecrets {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "envelope",
            "envelope_ref",
            "envelopeRef",
            "recipient_key_ids",
            "recipientKeyIds",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Envelope,
            EnvelopeRef,
            RecipientKeyIds,
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
                            "envelope" => Ok(GeneratedField::Envelope),
                            "envelopeRef" | "envelope_ref" => Ok(GeneratedField::EnvelopeRef),
                            "recipientKeyIds" | "recipient_key_ids" => Ok(GeneratedField::RecipientKeyIds),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EncryptedProviderSecrets;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.EncryptedProviderSecrets")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EncryptedProviderSecrets, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut envelope__ = None;
                let mut envelope_ref__ = None;
                let mut recipient_key_ids__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Envelope => {
                            if envelope__.is_some() {
                                return Err(serde::de::Error::duplicate_field("envelope"));
                            }
                            envelope__ = map_.next_value()?;
                        }
                        GeneratedField::EnvelopeRef => {
                            if envelope_ref__.is_some() {
                                return Err(serde::de::Error::duplicate_field("envelopeRef"));
                            }
                            envelope_ref__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecipientKeyIds => {
                            if recipient_key_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientKeyIds"));
                            }
                            recipient_key_ids__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(EncryptedProviderSecrets {
                    envelope: envelope__,
                    envelope_ref: envelope_ref__.unwrap_or_default(),
                    recipient_key_ids: recipient_key_ids__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.EncryptedProviderSecrets", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for IdentityRequirement {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.min_score != 0 {
            len += 1;
        }
        if !self.required_status.is_empty() {
            len += 1;
        }
        if self.require_verified_email {
            len += 1;
        }
        if self.require_verified_domain {
            len += 1;
        }
        if self.require_mfa {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.IdentityRequirement", len)?;
        if self.min_score != 0 {
            struct_ser.serialize_field("minScore", &self.min_score)?;
        }
        if !self.required_status.is_empty() {
            struct_ser.serialize_field("requiredStatus", &self.required_status)?;
        }
        if self.require_verified_email {
            struct_ser.serialize_field("requireVerifiedEmail", &self.require_verified_email)?;
        }
        if self.require_verified_domain {
            struct_ser.serialize_field("requireVerifiedDomain", &self.require_verified_domain)?;
        }
        if self.require_mfa {
            struct_ser.serialize_field("requireMfa", &self.require_mfa)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for IdentityRequirement {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "min_score",
            "minScore",
            "required_status",
            "requiredStatus",
            "require_verified_email",
            "requireVerifiedEmail",
            "require_verified_domain",
            "requireVerifiedDomain",
            "require_mfa",
            "requireMfa",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MinScore,
            RequiredStatus,
            RequireVerifiedEmail,
            RequireVerifiedDomain,
            RequireMfa,
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
                            "minScore" | "min_score" => Ok(GeneratedField::MinScore),
                            "requiredStatus" | "required_status" => Ok(GeneratedField::RequiredStatus),
                            "requireVerifiedEmail" | "require_verified_email" => Ok(GeneratedField::RequireVerifiedEmail),
                            "requireVerifiedDomain" | "require_verified_domain" => Ok(GeneratedField::RequireVerifiedDomain),
                            "requireMfa" | "require_mfa" => Ok(GeneratedField::RequireMfa),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IdentityRequirement;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.IdentityRequirement")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<IdentityRequirement, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut min_score__ = None;
                let mut required_status__ = None;
                let mut require_verified_email__ = None;
                let mut require_verified_domain__ = None;
                let mut require_mfa__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MinScore => {
                            if min_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minScore"));
                            }
                            min_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequiredStatus => {
                            if required_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredStatus"));
                            }
                            required_status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireVerifiedEmail => {
                            if require_verified_email__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireVerifiedEmail"));
                            }
                            require_verified_email__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireVerifiedDomain => {
                            if require_verified_domain__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireVerifiedDomain"));
                            }
                            require_verified_domain__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireMfa => {
                            if require_mfa__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireMfa"));
                            }
                            require_mfa__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(IdentityRequirement {
                    min_score: min_score__.unwrap_or_default(),
                    required_status: required_status__.unwrap_or_default(),
                    require_verified_email: require_verified_email__.unwrap_or_default(),
                    require_verified_domain: require_verified_domain__.unwrap_or_default(),
                    require_mfa: require_mfa__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.IdentityRequirement", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAcceptBid {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.customer.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if !self.bid_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgAcceptBid", len)?;
        if !self.customer.is_empty() {
            struct_ser.serialize_field("customer", &self.customer)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if !self.bid_id.is_empty() {
            struct_ser.serialize_field("bidId", &self.bid_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAcceptBid {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "customer",
            "order_id",
            "orderId",
            "bid_id",
            "bidId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Customer,
            OrderId,
            BidId,
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
                            "customer" => Ok(GeneratedField::Customer),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "bidId" | "bid_id" => Ok(GeneratedField::BidId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAcceptBid;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgAcceptBid")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAcceptBid, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut customer__ = None;
                let mut order_id__ = None;
                let mut bid_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Customer => {
                            if customer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customer"));
                            }
                            customer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BidId => {
                            if bid_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bidId"));
                            }
                            bid_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgAcceptBid {
                    customer: customer__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    bid_id: bid_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgAcceptBid", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAcceptBidResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.allocation_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgAcceptBidResponse", len)?;
        if !self.allocation_id.is_empty() {
            struct_ser.serialize_field("allocationId", &self.allocation_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAcceptBidResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "allocation_id",
            "allocationId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AllocationId,
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
                            "allocationId" | "allocation_id" => Ok(GeneratedField::AllocationId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAcceptBidResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgAcceptBidResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAcceptBidResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut allocation_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AllocationId => {
                            if allocation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocationId"));
                            }
                            allocation_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgAcceptBidResponse {
                    allocation_id: allocation_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgAcceptBidResponse", FIELDS, GeneratedVisitor)
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
        if !self.provider.is_empty() {
            len += 1;
        }
        if self.offering.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgCreateOffering", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if let Some(v) = self.offering.as_ref() {
            struct_ser.serialize_field("offering", v)?;
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
            "provider",
            "offering",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
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
                            "provider" => Ok(GeneratedField::Provider),
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
            type Value = MsgCreateOffering;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgCreateOffering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateOffering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut offering__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Offering => {
                            if offering__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offering"));
                            }
                            offering__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgCreateOffering {
                    provider: provider__.unwrap_or_default(),
                    offering: offering__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgCreateOffering", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgCreateOfferingResponse", len)?;
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
                formatter.write_str("struct virtengine.marketplace.v1.MsgCreateOfferingResponse")
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
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgCreateOfferingResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeactivateOffering {
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
        if !self.offering_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgDeactivateOffering", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeactivateOffering {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "offering_id",
            "offeringId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
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
                            "provider" => Ok(GeneratedField::Provider),
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
            type Value = MsgDeactivateOffering;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgDeactivateOffering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeactivateOffering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut offering_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDeactivateOffering {
                    provider: provider__.unwrap_or_default(),
                    offering_id: offering_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgDeactivateOffering", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeactivateOfferingResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgDeactivateOfferingResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeactivateOfferingResponse {
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
            type Value = MsgDeactivateOfferingResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgDeactivateOfferingResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeactivateOfferingResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDeactivateOfferingResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgDeactivateOfferingResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgPauseAllocation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.customer.is_empty() {
            len += 1;
        }
        if !self.allocation_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgPauseAllocation", len)?;
        if !self.customer.is_empty() {
            struct_ser.serialize_field("customer", &self.customer)?;
        }
        if !self.allocation_id.is_empty() {
            struct_ser.serialize_field("allocationId", &self.allocation_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgPauseAllocation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "customer",
            "allocation_id",
            "allocationId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Customer,
            AllocationId,
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
                            "customer" => Ok(GeneratedField::Customer),
                            "allocationId" | "allocation_id" => Ok(GeneratedField::AllocationId),
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
            type Value = MsgPauseAllocation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgPauseAllocation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgPauseAllocation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut customer__ = None;
                let mut allocation_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Customer => {
                            if customer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customer"));
                            }
                            customer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllocationId => {
                            if allocation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocationId"));
                            }
                            allocation_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgPauseAllocation {
                    customer: customer__.unwrap_or_default(),
                    allocation_id: allocation_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgPauseAllocation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgPauseAllocationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgPauseAllocationResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgPauseAllocationResponse {
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
            type Value = MsgPauseAllocationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgPauseAllocationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgPauseAllocationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgPauseAllocationResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgPauseAllocationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResizeAllocation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.customer.is_empty() {
            len += 1;
        }
        if !self.allocation_id.is_empty() {
            len += 1;
        }
        if !self.resource_units.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgResizeAllocation", len)?;
        if !self.customer.is_empty() {
            struct_ser.serialize_field("customer", &self.customer)?;
        }
        if !self.allocation_id.is_empty() {
            struct_ser.serialize_field("allocationId", &self.allocation_id)?;
        }
        if !self.resource_units.is_empty() {
            struct_ser.serialize_field("resourceUnits", &self.resource_units)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResizeAllocation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "customer",
            "allocation_id",
            "allocationId",
            "resource_units",
            "resourceUnits",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Customer,
            AllocationId,
            ResourceUnits,
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
                            "customer" => Ok(GeneratedField::Customer),
                            "allocationId" | "allocation_id" => Ok(GeneratedField::AllocationId),
                            "resourceUnits" | "resource_units" => Ok(GeneratedField::ResourceUnits),
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
            type Value = MsgResizeAllocation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgResizeAllocation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResizeAllocation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut customer__ = None;
                let mut allocation_id__ = None;
                let mut resource_units__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Customer => {
                            if customer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customer"));
                            }
                            customer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllocationId => {
                            if allocation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocationId"));
                            }
                            allocation_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResourceUnits => {
                            if resource_units__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resourceUnits"));
                            }
                            resource_units__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgResizeAllocation {
                    customer: customer__.unwrap_or_default(),
                    allocation_id: allocation_id__.unwrap_or_default(),
                    resource_units: resource_units__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgResizeAllocation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResizeAllocationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgResizeAllocationResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResizeAllocationResponse {
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
            type Value = MsgResizeAllocationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgResizeAllocationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResizeAllocationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgResizeAllocationResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgResizeAllocationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgTerminateAllocation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.customer.is_empty() {
            len += 1;
        }
        if !self.allocation_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgTerminateAllocation", len)?;
        if !self.customer.is_empty() {
            struct_ser.serialize_field("customer", &self.customer)?;
        }
        if !self.allocation_id.is_empty() {
            struct_ser.serialize_field("allocationId", &self.allocation_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgTerminateAllocation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "customer",
            "allocation_id",
            "allocationId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Customer,
            AllocationId,
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
                            "customer" => Ok(GeneratedField::Customer),
                            "allocationId" | "allocation_id" => Ok(GeneratedField::AllocationId),
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
            type Value = MsgTerminateAllocation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgTerminateAllocation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgTerminateAllocation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut customer__ = None;
                let mut allocation_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Customer => {
                            if customer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("customer"));
                            }
                            customer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllocationId => {
                            if allocation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocationId"));
                            }
                            allocation_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgTerminateAllocation {
                    customer: customer__.unwrap_or_default(),
                    allocation_id: allocation_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgTerminateAllocation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgTerminateAllocationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgTerminateAllocationResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgTerminateAllocationResponse {
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
            type Value = MsgTerminateAllocationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgTerminateAllocationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgTerminateAllocationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgTerminateAllocationResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgTerminateAllocationResponse", FIELDS, GeneratedVisitor)
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
        if !self.provider.is_empty() {
            len += 1;
        }
        if !self.offering_id.is_empty() {
            len += 1;
        }
        if self.updates.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgUpdateOffering", len)?;
        if !self.provider.is_empty() {
            struct_ser.serialize_field("provider", &self.provider)?;
        }
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if let Some(v) = self.updates.as_ref() {
            struct_ser.serialize_field("updates", v)?;
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
            "provider",
            "offering_id",
            "offeringId",
            "updates",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
            OfferingId,
            Updates,
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
                            "offeringId" | "offering_id" => Ok(GeneratedField::OfferingId),
                            "updates" => Ok(GeneratedField::Updates),
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
                formatter.write_str("struct virtengine.marketplace.v1.MsgUpdateOffering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateOffering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut offering_id__ = None;
                let mut updates__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Updates => {
                            if updates__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updates"));
                            }
                            updates__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgUpdateOffering {
                    provider: provider__.unwrap_or_default(),
                    offering_id: offering_id__.unwrap_or_default(),
                    updates: updates__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgUpdateOffering", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgUpdateOfferingResponse", len)?;
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
                formatter.write_str("struct virtengine.marketplace.v1.MsgUpdateOfferingResponse")
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
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgUpdateOfferingResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgWaldurCallback {
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
        if !self.callback_type.is_empty() {
            len += 1;
        }
        if !self.resource_id.is_empty() {
            len += 1;
        }
        if !self.status.is_empty() {
            len += 1;
        }
        if !self.payload.is_empty() {
            len += 1;
        }
        if !self.signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgWaldurCallback", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.callback_type.is_empty() {
            struct_ser.serialize_field("callbackType", &self.callback_type)?;
        }
        if !self.resource_id.is_empty() {
            struct_ser.serialize_field("resourceId", &self.resource_id)?;
        }
        if !self.status.is_empty() {
            struct_ser.serialize_field("status", &self.status)?;
        }
        if !self.payload.is_empty() {
            struct_ser.serialize_field("payload", &self.payload)?;
        }
        if !self.signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signature", pbjson::private::base64::encode(&self.signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgWaldurCallback {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "callback_type",
            "callbackType",
            "resource_id",
            "resourceId",
            "status",
            "payload",
            "signature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            CallbackType,
            ResourceId,
            Status,
            Payload,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "callbackType" | "callback_type" => Ok(GeneratedField::CallbackType),
                            "resourceId" | "resource_id" => Ok(GeneratedField::ResourceId),
                            "status" => Ok(GeneratedField::Status),
                            "payload" => Ok(GeneratedField::Payload),
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
            type Value = MsgWaldurCallback;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgWaldurCallback")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgWaldurCallback, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut callback_type__ = None;
                let mut resource_id__ = None;
                let mut status__ = None;
                let mut payload__ = None;
                let mut signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CallbackType => {
                            if callback_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("callbackType"));
                            }
                            callback_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResourceId => {
                            if resource_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resourceId"));
                            }
                            resource_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Payload => {
                            if payload__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payload"));
                            }
                            payload__ = Some(map_.next_value()?);
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
                Ok(MsgWaldurCallback {
                    sender: sender__.unwrap_or_default(),
                    callback_type: callback_type__.unwrap_or_default(),
                    resource_id: resource_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    payload: payload__.unwrap_or_default(),
                    signature: signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgWaldurCallback", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgWaldurCallbackResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.MsgWaldurCallbackResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgWaldurCallbackResponse {
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
            type Value = MsgWaldurCallbackResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.MsgWaldurCallbackResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgWaldurCallbackResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgWaldurCallbackResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.MsgWaldurCallbackResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Offering {
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
        if self.state != 0 {
            len += 1;
        }
        if self.category != 0 {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.version.is_empty() {
            len += 1;
        }
        if self.pricing.is_some() {
            len += 1;
        }
        if self.identity_requirement.is_some() {
            len += 1;
        }
        if self.require_mfa_for_orders {
            len += 1;
        }
        if !self.public_metadata.is_empty() {
            len += 1;
        }
        if self.encrypted_secrets.is_some() {
            len += 1;
        }
        if !self.specifications.is_empty() {
            len += 1;
        }
        if !self.tags.is_empty() {
            len += 1;
        }
        if !self.regions.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.activated_at.is_some() {
            len += 1;
        }
        if self.terminated_at.is_some() {
            len += 1;
        }
        if self.max_concurrent_orders != 0 {
            len += 1;
        }
        if self.total_order_count != 0 {
            len += 1;
        }
        if self.active_order_count != 0 {
            len += 1;
        }
        if !self.prices.is_empty() {
            len += 1;
        }
        if self.allow_bidding {
            len += 1;
        }
        if self.min_bid.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.Offering", len)?;
        if let Some(v) = self.id.as_ref() {
            struct_ser.serialize_field("id", v)?;
        }
        if self.state != 0 {
            let v = OfferingState::try_from(self.state)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if self.category != 0 {
            let v = OfferingCategory::try_from(self.category)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.category)))?;
            struct_ser.serialize_field("category", &v)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.version.is_empty() {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if let Some(v) = self.pricing.as_ref() {
            struct_ser.serialize_field("pricing", v)?;
        }
        if let Some(v) = self.identity_requirement.as_ref() {
            struct_ser.serialize_field("identityRequirement", v)?;
        }
        if self.require_mfa_for_orders {
            struct_ser.serialize_field("requireMfaForOrders", &self.require_mfa_for_orders)?;
        }
        if !self.public_metadata.is_empty() {
            struct_ser.serialize_field("publicMetadata", &self.public_metadata)?;
        }
        if let Some(v) = self.encrypted_secrets.as_ref() {
            struct_ser.serialize_field("encryptedSecrets", v)?;
        }
        if !self.specifications.is_empty() {
            struct_ser.serialize_field("specifications", &self.specifications)?;
        }
        if !self.tags.is_empty() {
            struct_ser.serialize_field("tags", &self.tags)?;
        }
        if !self.regions.is_empty() {
            struct_ser.serialize_field("regions", &self.regions)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if let Some(v) = self.activated_at.as_ref() {
            struct_ser.serialize_field("activatedAt", v)?;
        }
        if let Some(v) = self.terminated_at.as_ref() {
            struct_ser.serialize_field("terminatedAt", v)?;
        }
        if self.max_concurrent_orders != 0 {
            struct_ser.serialize_field("maxConcurrentOrders", &self.max_concurrent_orders)?;
        }
        if self.total_order_count != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("totalOrderCount", ToString::to_string(&self.total_order_count).as_str())?;
        }
        if self.active_order_count != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activeOrderCount", ToString::to_string(&self.active_order_count).as_str())?;
        }
        if !self.prices.is_empty() {
            struct_ser.serialize_field("prices", &self.prices)?;
        }
        if self.allow_bidding {
            struct_ser.serialize_field("allowBidding", &self.allow_bidding)?;
        }
        if let Some(v) = self.min_bid.as_ref() {
            struct_ser.serialize_field("minBid", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Offering {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "state",
            "category",
            "name",
            "description",
            "version",
            "pricing",
            "identity_requirement",
            "identityRequirement",
            "require_mfa_for_orders",
            "requireMfaForOrders",
            "public_metadata",
            "publicMetadata",
            "encrypted_secrets",
            "encryptedSecrets",
            "specifications",
            "tags",
            "regions",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "activated_at",
            "activatedAt",
            "terminated_at",
            "terminatedAt",
            "max_concurrent_orders",
            "maxConcurrentOrders",
            "total_order_count",
            "totalOrderCount",
            "active_order_count",
            "activeOrderCount",
            "prices",
            "allow_bidding",
            "allowBidding",
            "min_bid",
            "minBid",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            State,
            Category,
            Name,
            Description,
            Version,
            Pricing,
            IdentityRequirement,
            RequireMfaForOrders,
            PublicMetadata,
            EncryptedSecrets,
            Specifications,
            Tags,
            Regions,
            CreatedAt,
            UpdatedAt,
            ActivatedAt,
            TerminatedAt,
            MaxConcurrentOrders,
            TotalOrderCount,
            ActiveOrderCount,
            Prices,
            AllowBidding,
            MinBid,
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
                            "category" => Ok(GeneratedField::Category),
                            "name" => Ok(GeneratedField::Name),
                            "description" => Ok(GeneratedField::Description),
                            "version" => Ok(GeneratedField::Version),
                            "pricing" => Ok(GeneratedField::Pricing),
                            "identityRequirement" | "identity_requirement" => Ok(GeneratedField::IdentityRequirement),
                            "requireMfaForOrders" | "require_mfa_for_orders" => Ok(GeneratedField::RequireMfaForOrders),
                            "publicMetadata" | "public_metadata" => Ok(GeneratedField::PublicMetadata),
                            "encryptedSecrets" | "encrypted_secrets" => Ok(GeneratedField::EncryptedSecrets),
                            "specifications" => Ok(GeneratedField::Specifications),
                            "tags" => Ok(GeneratedField::Tags),
                            "regions" => Ok(GeneratedField::Regions),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "activatedAt" | "activated_at" => Ok(GeneratedField::ActivatedAt),
                            "terminatedAt" | "terminated_at" => Ok(GeneratedField::TerminatedAt),
                            "maxConcurrentOrders" | "max_concurrent_orders" => Ok(GeneratedField::MaxConcurrentOrders),
                            "totalOrderCount" | "total_order_count" => Ok(GeneratedField::TotalOrderCount),
                            "activeOrderCount" | "active_order_count" => Ok(GeneratedField::ActiveOrderCount),
                            "prices" => Ok(GeneratedField::Prices),
                            "allowBidding" | "allow_bidding" => Ok(GeneratedField::AllowBidding),
                            "minBid" | "min_bid" => Ok(GeneratedField::MinBid),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Offering;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.Offering")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Offering, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut state__ = None;
                let mut category__ = None;
                let mut name__ = None;
                let mut description__ = None;
                let mut version__ = None;
                let mut pricing__ = None;
                let mut identity_requirement__ = None;
                let mut require_mfa_for_orders__ = None;
                let mut public_metadata__ = None;
                let mut encrypted_secrets__ = None;
                let mut specifications__ = None;
                let mut tags__ = None;
                let mut regions__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut activated_at__ = None;
                let mut terminated_at__ = None;
                let mut max_concurrent_orders__ = None;
                let mut total_order_count__ = None;
                let mut active_order_count__ = None;
                let mut prices__ = None;
                let mut allow_bidding__ = None;
                let mut min_bid__ = None;
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
                            state__ = Some(map_.next_value::<OfferingState>()? as i32);
                        }
                        GeneratedField::Category => {
                            if category__.is_some() {
                                return Err(serde::de::Error::duplicate_field("category"));
                            }
                            category__ = Some(map_.next_value::<OfferingCategory>()? as i32);
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
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pricing => {
                            if pricing__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pricing"));
                            }
                            pricing__ = map_.next_value()?;
                        }
                        GeneratedField::IdentityRequirement => {
                            if identity_requirement__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identityRequirement"));
                            }
                            identity_requirement__ = map_.next_value()?;
                        }
                        GeneratedField::RequireMfaForOrders => {
                            if require_mfa_for_orders__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireMfaForOrders"));
                            }
                            require_mfa_for_orders__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicMetadata => {
                            if public_metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicMetadata"));
                            }
                            public_metadata__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                        GeneratedField::EncryptedSecrets => {
                            if encrypted_secrets__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedSecrets"));
                            }
                            encrypted_secrets__ = map_.next_value()?;
                        }
                        GeneratedField::Specifications => {
                            if specifications__.is_some() {
                                return Err(serde::de::Error::duplicate_field("specifications"));
                            }
                            specifications__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                        GeneratedField::Tags => {
                            if tags__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tags"));
                            }
                            tags__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Regions => {
                            if regions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("regions"));
                            }
                            regions__ = Some(map_.next_value()?);
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
                        GeneratedField::ActivatedAt => {
                            if activated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activatedAt"));
                            }
                            activated_at__ = map_.next_value()?;
                        }
                        GeneratedField::TerminatedAt => {
                            if terminated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("terminatedAt"));
                            }
                            terminated_at__ = map_.next_value()?;
                        }
                        GeneratedField::MaxConcurrentOrders => {
                            if max_concurrent_orders__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxConcurrentOrders"));
                            }
                            max_concurrent_orders__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TotalOrderCount => {
                            if total_order_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalOrderCount"));
                            }
                            total_order_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ActiveOrderCount => {
                            if active_order_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activeOrderCount"));
                            }
                            active_order_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Prices => {
                            if prices__.is_some() {
                                return Err(serde::de::Error::duplicate_field("prices"));
                            }
                            prices__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowBidding => {
                            if allow_bidding__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowBidding"));
                            }
                            allow_bidding__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinBid => {
                            if min_bid__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minBid"));
                            }
                            min_bid__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Offering {
                    id: id__,
                    state: state__.unwrap_or_default(),
                    category: category__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    version: version__.unwrap_or_default(),
                    pricing: pricing__,
                    identity_requirement: identity_requirement__,
                    require_mfa_for_orders: require_mfa_for_orders__.unwrap_or_default(),
                    public_metadata: public_metadata__.unwrap_or_default(),
                    encrypted_secrets: encrypted_secrets__,
                    specifications: specifications__.unwrap_or_default(),
                    tags: tags__.unwrap_or_default(),
                    regions: regions__.unwrap_or_default(),
                    created_at: created_at__,
                    updated_at: updated_at__,
                    activated_at: activated_at__,
                    terminated_at: terminated_at__,
                    max_concurrent_orders: max_concurrent_orders__.unwrap_or_default(),
                    total_order_count: total_order_count__.unwrap_or_default(),
                    active_order_count: active_order_count__.unwrap_or_default(),
                    prices: prices__.unwrap_or_default(),
                    allow_bidding: allow_bidding__.unwrap_or_default(),
                    min_bid: min_bid__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.Offering", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for OfferingCategory {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "OFFERING_CATEGORY_UNSPECIFIED",
            Self::Compute => "OFFERING_CATEGORY_COMPUTE",
            Self::Storage => "OFFERING_CATEGORY_STORAGE",
            Self::Network => "OFFERING_CATEGORY_NETWORK",
            Self::Hpc => "OFFERING_CATEGORY_HPC",
            Self::Gpu => "OFFERING_CATEGORY_GPU",
            Self::Ml => "OFFERING_CATEGORY_ML",
            Self::Other => "OFFERING_CATEGORY_OTHER",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for OfferingCategory {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "OFFERING_CATEGORY_UNSPECIFIED",
            "OFFERING_CATEGORY_COMPUTE",
            "OFFERING_CATEGORY_STORAGE",
            "OFFERING_CATEGORY_NETWORK",
            "OFFERING_CATEGORY_HPC",
            "OFFERING_CATEGORY_GPU",
            "OFFERING_CATEGORY_ML",
            "OFFERING_CATEGORY_OTHER",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = OfferingCategory;

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
                    "OFFERING_CATEGORY_UNSPECIFIED" => Ok(OfferingCategory::Unspecified),
                    "OFFERING_CATEGORY_COMPUTE" => Ok(OfferingCategory::Compute),
                    "OFFERING_CATEGORY_STORAGE" => Ok(OfferingCategory::Storage),
                    "OFFERING_CATEGORY_NETWORK" => Ok(OfferingCategory::Network),
                    "OFFERING_CATEGORY_HPC" => Ok(OfferingCategory::Hpc),
                    "OFFERING_CATEGORY_GPU" => Ok(OfferingCategory::Gpu),
                    "OFFERING_CATEGORY_ML" => Ok(OfferingCategory::Ml),
                    "OFFERING_CATEGORY_OTHER" => Ok(OfferingCategory::Other),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for OfferingId {
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
        if self.sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.OfferingID", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if self.sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("sequence", ToString::to_string(&self.sequence).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for OfferingId {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "sequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
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
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
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
            type Value = OfferingId;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.OfferingID")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<OfferingId, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
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
                Ok(OfferingId {
                    provider_address: provider_address__.unwrap_or_default(),
                    sequence: sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.OfferingID", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for OfferingState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "OFFERING_STATE_UNSPECIFIED",
            Self::Active => "OFFERING_STATE_ACTIVE",
            Self::Paused => "OFFERING_STATE_PAUSED",
            Self::Suspended => "OFFERING_STATE_SUSPENDED",
            Self::Deprecated => "OFFERING_STATE_DEPRECATED",
            Self::Terminated => "OFFERING_STATE_TERMINATED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for OfferingState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "OFFERING_STATE_UNSPECIFIED",
            "OFFERING_STATE_ACTIVE",
            "OFFERING_STATE_PAUSED",
            "OFFERING_STATE_SUSPENDED",
            "OFFERING_STATE_DEPRECATED",
            "OFFERING_STATE_TERMINATED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = OfferingState;

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
                    "OFFERING_STATE_UNSPECIFIED" => Ok(OfferingState::Unspecified),
                    "OFFERING_STATE_ACTIVE" => Ok(OfferingState::Active),
                    "OFFERING_STATE_PAUSED" => Ok(OfferingState::Paused),
                    "OFFERING_STATE_SUSPENDED" => Ok(OfferingState::Suspended),
                    "OFFERING_STATE_DEPRECATED" => Ok(OfferingState::Deprecated),
                    "OFFERING_STATE_TERMINATED" => Ok(OfferingState::Terminated),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for PriceComponent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.resource_type.is_empty() {
            len += 1;
        }
        if !self.unit.is_empty() {
            len += 1;
        }
        if self.price.is_some() {
            len += 1;
        }
        if !self.usd_reference.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.PriceComponent", len)?;
        if !self.resource_type.is_empty() {
            struct_ser.serialize_field("resourceType", &self.resource_type)?;
        }
        if !self.unit.is_empty() {
            struct_ser.serialize_field("unit", &self.unit)?;
        }
        if let Some(v) = self.price.as_ref() {
            struct_ser.serialize_field("price", v)?;
        }
        if !self.usd_reference.is_empty() {
            struct_ser.serialize_field("usdReference", &self.usd_reference)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PriceComponent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "resource_type",
            "resourceType",
            "unit",
            "price",
            "usd_reference",
            "usdReference",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ResourceType,
            Unit,
            Price,
            UsdReference,
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
                            "resourceType" | "resource_type" => Ok(GeneratedField::ResourceType),
                            "unit" => Ok(GeneratedField::Unit),
                            "price" => Ok(GeneratedField::Price),
                            "usdReference" | "usd_reference" => Ok(GeneratedField::UsdReference),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PriceComponent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.PriceComponent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PriceComponent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut resource_type__ = None;
                let mut unit__ = None;
                let mut price__ = None;
                let mut usd_reference__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ResourceType => {
                            if resource_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resourceType"));
                            }
                            resource_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Unit => {
                            if unit__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unit"));
                            }
                            unit__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Price => {
                            if price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("price"));
                            }
                            price__ = map_.next_value()?;
                        }
                        GeneratedField::UsdReference => {
                            if usd_reference__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usdReference"));
                            }
                            usd_reference__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(PriceComponent {
                    resource_type: resource_type__.unwrap_or_default(),
                    unit: unit__.unwrap_or_default(),
                    price: price__,
                    usd_reference: usd_reference__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.PriceComponent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PricingInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.model != 0 {
            len += 1;
        }
        if self.base_price != 0 {
            len += 1;
        }
        if !self.currency.is_empty() {
            len += 1;
        }
        if !self.usage_rates.is_empty() {
            len += 1;
        }
        if self.minimum_commitment != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.PricingInfo", len)?;
        if self.model != 0 {
            let v = PricingModel::try_from(self.model)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.model)))?;
            struct_ser.serialize_field("model", &v)?;
        }
        if self.base_price != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("basePrice", ToString::to_string(&self.base_price).as_str())?;
        }
        if !self.currency.is_empty() {
            struct_ser.serialize_field("currency", &self.currency)?;
        }
        if !self.usage_rates.is_empty() {
            let v: std::collections::HashMap<_, _> = self.usage_rates.iter()
                .map(|(k, v)| (k, v.to_string())).collect();
            struct_ser.serialize_field("usageRates", &v)?;
        }
        if self.minimum_commitment != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("minimumCommitment", ToString::to_string(&self.minimum_commitment).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PricingInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "model",
            "base_price",
            "basePrice",
            "currency",
            "usage_rates",
            "usageRates",
            "minimum_commitment",
            "minimumCommitment",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Model,
            BasePrice,
            Currency,
            UsageRates,
            MinimumCommitment,
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
                            "model" => Ok(GeneratedField::Model),
                            "basePrice" | "base_price" => Ok(GeneratedField::BasePrice),
                            "currency" => Ok(GeneratedField::Currency),
                            "usageRates" | "usage_rates" => Ok(GeneratedField::UsageRates),
                            "minimumCommitment" | "minimum_commitment" => Ok(GeneratedField::MinimumCommitment),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PricingInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.PricingInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PricingInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut model__ = None;
                let mut base_price__ = None;
                let mut currency__ = None;
                let mut usage_rates__ = None;
                let mut minimum_commitment__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Model => {
                            if model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("model"));
                            }
                            model__ = Some(map_.next_value::<PricingModel>()? as i32);
                        }
                        GeneratedField::BasePrice => {
                            if base_price__.is_some() {
                                return Err(serde::de::Error::duplicate_field("basePrice"));
                            }
                            base_price__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Currency => {
                            if currency__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currency"));
                            }
                            currency__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UsageRates => {
                            if usage_rates__.is_some() {
                                return Err(serde::de::Error::duplicate_field("usageRates"));
                            }
                            usage_rates__ = Some(
                                map_.next_value::<std::collections::HashMap<_, ::pbjson::private::NumberDeserialize<u64>>>()?
                                    .into_iter().map(|(k,v)| (k, v.0)).collect()
                            );
                        }
                        GeneratedField::MinimumCommitment => {
                            if minimum_commitment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minimumCommitment"));
                            }
                            minimum_commitment__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(PricingInfo {
                    model: model__.unwrap_or_default(),
                    base_price: base_price__.unwrap_or_default(),
                    currency: currency__.unwrap_or_default(),
                    usage_rates: usage_rates__.unwrap_or_default(),
                    minimum_commitment: minimum_commitment__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.PricingInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PricingModel {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "PRICING_MODEL_UNSPECIFIED",
            Self::Hourly => "PRICING_MODEL_HOURLY",
            Self::Daily => "PRICING_MODEL_DAILY",
            Self::Monthly => "PRICING_MODEL_MONTHLY",
            Self::UsageBased => "PRICING_MODEL_USAGE_BASED",
            Self::Fixed => "PRICING_MODEL_FIXED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for PricingModel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "PRICING_MODEL_UNSPECIFIED",
            "PRICING_MODEL_HOURLY",
            "PRICING_MODEL_DAILY",
            "PRICING_MODEL_MONTHLY",
            "PRICING_MODEL_USAGE_BASED",
            "PRICING_MODEL_FIXED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PricingModel;

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
                    "PRICING_MODEL_UNSPECIFIED" => Ok(PricingModel::Unspecified),
                    "PRICING_MODEL_HOURLY" => Ok(PricingModel::Hourly),
                    "PRICING_MODEL_DAILY" => Ok(PricingModel::Daily),
                    "PRICING_MODEL_MONTHLY" => Ok(PricingModel::Monthly),
                    "PRICING_MODEL_USAGE_BASED" => Ok(PricingModel::UsageBased),
                    "PRICING_MODEL_FIXED" => Ok(PricingModel::Fixed),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAllocationsByCustomerRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.QueryAllocationsByCustomerRequest", len)?;
        if !self.customer_address.is_empty() {
            struct_ser.serialize_field("customerAddress", &self.customer_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAllocationsByCustomerRequest {
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
            type Value = QueryAllocationsByCustomerRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.QueryAllocationsByCustomerRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAllocationsByCustomerRequest, V::Error>
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
                Ok(QueryAllocationsByCustomerRequest {
                    customer_address: customer_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.QueryAllocationsByCustomerRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAllocationsByProviderRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.QueryAllocationsByProviderRequest", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAllocationsByProviderRequest {
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
            type Value = QueryAllocationsByProviderRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.QueryAllocationsByProviderRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAllocationsByProviderRequest, V::Error>
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
                Ok(QueryAllocationsByProviderRequest {
                    provider_address: provider_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.QueryAllocationsByProviderRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAllocationsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.allocations.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.QueryAllocationsResponse", len)?;
        if !self.allocations.is_empty() {
            struct_ser.serialize_field("allocations", &self.allocations)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAllocationsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "allocations",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Allocations,
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
                            "allocations" => Ok(GeneratedField::Allocations),
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
            type Value = QueryAllocationsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.QueryAllocationsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAllocationsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut allocations__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Allocations => {
                            if allocations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allocations"));
                            }
                            allocations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAllocationsResponse {
                    allocations: allocations__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.QueryAllocationsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingPriceRequest {
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
        if !self.resource_units.is_empty() {
            len += 1;
        }
        if self.quantity != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.QueryOfferingPriceRequest", len)?;
        if !self.offering_id.is_empty() {
            struct_ser.serialize_field("offeringId", &self.offering_id)?;
        }
        if !self.resource_units.is_empty() {
            let v: std::collections::HashMap<_, _> = self.resource_units.iter()
                .map(|(k, v)| (k, v.to_string())).collect();
            struct_ser.serialize_field("resourceUnits", &v)?;
        }
        if self.quantity != 0 {
            struct_ser.serialize_field("quantity", &self.quantity)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingPriceRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "offering_id",
            "offeringId",
            "resource_units",
            "resourceUnits",
            "quantity",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            OfferingId,
            ResourceUnits,
            Quantity,
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
                            "resourceUnits" | "resource_units" => Ok(GeneratedField::ResourceUnits),
                            "quantity" => Ok(GeneratedField::Quantity),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryOfferingPriceRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.QueryOfferingPriceRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingPriceRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut offering_id__ = None;
                let mut resource_units__ = None;
                let mut quantity__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::OfferingId => {
                            if offering_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offeringId"));
                            }
                            offering_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ResourceUnits => {
                            if resource_units__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resourceUnits"));
                            }
                            resource_units__ = Some(
                                map_.next_value::<std::collections::HashMap<_, ::pbjson::private::NumberDeserialize<u64>>>()?
                                    .into_iter().map(|(k,v)| (k, v.0)).collect()
                            );
                        }
                        GeneratedField::Quantity => {
                            if quantity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("quantity"));
                            }
                            quantity__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryOfferingPriceRequest {
                    offering_id: offering_id__.unwrap_or_default(),
                    resource_units: resource_units__.unwrap_or_default(),
                    quantity: quantity__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.QueryOfferingPriceRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryOfferingPriceResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.total.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.QueryOfferingPriceResponse", len)?;
        if let Some(v) = self.total.as_ref() {
            struct_ser.serialize_field("total", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryOfferingPriceResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "total",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Total,
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
                            "total" => Ok(GeneratedField::Total),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryOfferingPriceResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.QueryOfferingPriceResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryOfferingPriceResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut total__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Total => {
                            if total__.is_some() {
                                return Err(serde::de::Error::duplicate_field("total"));
                            }
                            total__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryOfferingPriceResponse {
                    total: total__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.QueryOfferingPriceResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ResourceUnit {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.resource_type.is_empty() {
            len += 1;
        }
        if self.units != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.marketplace.v1.ResourceUnit", len)?;
        if !self.resource_type.is_empty() {
            struct_ser.serialize_field("resourceType", &self.resource_type)?;
        }
        if self.units != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("units", ToString::to_string(&self.units).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ResourceUnit {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "resource_type",
            "resourceType",
            "units",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ResourceType,
            Units,
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
                            "resourceType" | "resource_type" => Ok(GeneratedField::ResourceType),
                            "units" => Ok(GeneratedField::Units),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ResourceUnit;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.marketplace.v1.ResourceUnit")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ResourceUnit, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut resource_type__ = None;
                let mut units__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ResourceType => {
                            if resource_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resourceType"));
                            }
                            resource_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Units => {
                            if units__.is_some() {
                                return Err(serde::de::Error::duplicate_field("units"));
                            }
                            units__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ResourceUnit {
                    resource_type: resource_type__.unwrap_or_default(),
                    units: units__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.marketplace.v1.ResourceUnit", FIELDS, GeneratedVisitor)
    }
}

// @generated
impl serde::Serialize for Delegation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.shares.is_empty() {
            len += 1;
        }
        if !self.initial_amount.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.Delegation", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.shares.is_empty() {
            struct_ser.serialize_field("shares", &self.shares)?;
        }
        if !self.initial_amount.is_empty() {
            struct_ser.serialize_field("initialAmount", &self.initial_amount)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Delegation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "validator_address",
            "validatorAddress",
            "shares",
            "initial_amount",
            "initialAmount",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "height",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
            ValidatorAddress,
            Shares,
            InitialAmount,
            CreatedAt,
            UpdatedAt,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "shares" => Ok(GeneratedField::Shares),
                            "initialAmount" | "initial_amount" => Ok(GeneratedField::InitialAmount),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
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
            type Value = Delegation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.Delegation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Delegation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut validator_address__ = None;
                let mut shares__ = None;
                let mut initial_amount__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Shares => {
                            if shares__.is_some() {
                                return Err(serde::de::Error::duplicate_field("shares"));
                            }
                            shares__ = Some(map_.next_value()?);
                        }
                        GeneratedField::InitialAmount => {
                            if initial_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("initialAmount"));
                            }
                            initial_amount__ = Some(map_.next_value()?);
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
                Ok(Delegation {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    shares: shares__.unwrap_or_default(),
                    initial_amount: initial_amount__.unwrap_or_default(),
                    created_at: created_at__,
                    updated_at: updated_at__,
                    height: height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.Delegation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for DelegationStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "DELEGATION_STATUS_UNSPECIFIED",
            Self::Active => "DELEGATION_STATUS_ACTIVE",
            Self::Unbonding => "DELEGATION_STATUS_UNBONDING",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for DelegationStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "DELEGATION_STATUS_UNSPECIFIED",
            "DELEGATION_STATUS_ACTIVE",
            "DELEGATION_STATUS_UNBONDING",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DelegationStatus;

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
                    "DELEGATION_STATUS_UNSPECIFIED" => Ok(DelegationStatus::Unspecified),
                    "DELEGATION_STATUS_ACTIVE" => Ok(DelegationStatus::Active),
                    "DELEGATION_STATUS_UNBONDING" => Ok(DelegationStatus::Unbonding),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for DelegatorReward {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.epoch_number != 0 {
            len += 1;
        }
        if !self.reward.is_empty() {
            len += 1;
        }
        if !self.shares_at_epoch.is_empty() {
            len += 1;
        }
        if !self.validator_total_shares_at_epoch.is_empty() {
            len += 1;
        }
        if self.calculated_at.is_some() {
            len += 1;
        }
        if self.claimed {
            len += 1;
        }
        if self.claimed_at.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.DelegatorReward", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.epoch_number != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if !self.reward.is_empty() {
            struct_ser.serialize_field("reward", &self.reward)?;
        }
        if !self.shares_at_epoch.is_empty() {
            struct_ser.serialize_field("sharesAtEpoch", &self.shares_at_epoch)?;
        }
        if !self.validator_total_shares_at_epoch.is_empty() {
            struct_ser.serialize_field("validatorTotalSharesAtEpoch", &self.validator_total_shares_at_epoch)?;
        }
        if let Some(v) = self.calculated_at.as_ref() {
            struct_ser.serialize_field("calculatedAt", v)?;
        }
        if self.claimed {
            struct_ser.serialize_field("claimed", &self.claimed)?;
        }
        if let Some(v) = self.claimed_at.as_ref() {
            struct_ser.serialize_field("claimedAt", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DelegatorReward {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "validator_address",
            "validatorAddress",
            "epoch_number",
            "epochNumber",
            "reward",
            "shares_at_epoch",
            "sharesAtEpoch",
            "validator_total_shares_at_epoch",
            "validatorTotalSharesAtEpoch",
            "calculated_at",
            "calculatedAt",
            "claimed",
            "claimed_at",
            "claimedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
            ValidatorAddress,
            EpochNumber,
            Reward,
            SharesAtEpoch,
            ValidatorTotalSharesAtEpoch,
            CalculatedAt,
            Claimed,
            ClaimedAt,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "epochNumber" | "epoch_number" => Ok(GeneratedField::EpochNumber),
                            "reward" => Ok(GeneratedField::Reward),
                            "sharesAtEpoch" | "shares_at_epoch" => Ok(GeneratedField::SharesAtEpoch),
                            "validatorTotalSharesAtEpoch" | "validator_total_shares_at_epoch" => Ok(GeneratedField::ValidatorTotalSharesAtEpoch),
                            "calculatedAt" | "calculated_at" => Ok(GeneratedField::CalculatedAt),
                            "claimed" => Ok(GeneratedField::Claimed),
                            "claimedAt" | "claimed_at" => Ok(GeneratedField::ClaimedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DelegatorReward;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.DelegatorReward")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DelegatorReward, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut validator_address__ = None;
                let mut epoch_number__ = None;
                let mut reward__ = None;
                let mut shares_at_epoch__ = None;
                let mut validator_total_shares_at_epoch__ = None;
                let mut calculated_at__ = None;
                let mut claimed__ = None;
                let mut claimed_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EpochNumber => {
                            if epoch_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNumber"));
                            }
                            epoch_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Reward => {
                            if reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reward"));
                            }
                            reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SharesAtEpoch => {
                            if shares_at_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sharesAtEpoch"));
                            }
                            shares_at_epoch__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorTotalSharesAtEpoch => {
                            if validator_total_shares_at_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorTotalSharesAtEpoch"));
                            }
                            validator_total_shares_at_epoch__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CalculatedAt => {
                            if calculated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("calculatedAt"));
                            }
                            calculated_at__ = map_.next_value()?;
                        }
                        GeneratedField::Claimed => {
                            if claimed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("claimed"));
                            }
                            claimed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClaimedAt => {
                            if claimed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("claimedAt"));
                            }
                            claimed_at__ = map_.next_value()?;
                        }
                    }
                }
                Ok(DelegatorReward {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    epoch_number: epoch_number__.unwrap_or_default(),
                    reward: reward__.unwrap_or_default(),
                    shares_at_epoch: shares_at_epoch__.unwrap_or_default(),
                    validator_total_shares_at_epoch: validator_total_shares_at_epoch__.unwrap_or_default(),
                    calculated_at: calculated_at__,
                    claimed: claimed__.unwrap_or_default(),
                    claimed_at: claimed_at__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.DelegatorReward", FIELDS, GeneratedVisitor)
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
        if !self.delegations.is_empty() {
            len += 1;
        }
        if !self.unbonding_delegations.is_empty() {
            len += 1;
        }
        if !self.redelegations.is_empty() {
            len += 1;
        }
        if !self.validator_shares.is_empty() {
            len += 1;
        }
        if !self.delegator_rewards.is_empty() {
            len += 1;
        }
        if self.delegation_sequence != 0 {
            len += 1;
        }
        if self.unbonding_sequence != 0 {
            len += 1;
        }
        if self.redelegation_sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.delegations.is_empty() {
            struct_ser.serialize_field("delegations", &self.delegations)?;
        }
        if !self.unbonding_delegations.is_empty() {
            struct_ser.serialize_field("unbondingDelegations", &self.unbonding_delegations)?;
        }
        if !self.redelegations.is_empty() {
            struct_ser.serialize_field("redelegations", &self.redelegations)?;
        }
        if !self.validator_shares.is_empty() {
            struct_ser.serialize_field("validatorShares", &self.validator_shares)?;
        }
        if !self.delegator_rewards.is_empty() {
            struct_ser.serialize_field("delegatorRewards", &self.delegator_rewards)?;
        }
        if self.delegation_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("delegationSequence", ToString::to_string(&self.delegation_sequence).as_str())?;
        }
        if self.unbonding_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("unbondingSequence", ToString::to_string(&self.unbonding_sequence).as_str())?;
        }
        if self.redelegation_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("redelegationSequence", ToString::to_string(&self.redelegation_sequence).as_str())?;
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
            "delegations",
            "unbonding_delegations",
            "unbondingDelegations",
            "redelegations",
            "validator_shares",
            "validatorShares",
            "delegator_rewards",
            "delegatorRewards",
            "delegation_sequence",
            "delegationSequence",
            "unbonding_sequence",
            "unbondingSequence",
            "redelegation_sequence",
            "redelegationSequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            Delegations,
            UnbondingDelegations,
            Redelegations,
            ValidatorShares,
            DelegatorRewards,
            DelegationSequence,
            UnbondingSequence,
            RedelegationSequence,
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
                            "delegations" => Ok(GeneratedField::Delegations),
                            "unbondingDelegations" | "unbonding_delegations" => Ok(GeneratedField::UnbondingDelegations),
                            "redelegations" => Ok(GeneratedField::Redelegations),
                            "validatorShares" | "validator_shares" => Ok(GeneratedField::ValidatorShares),
                            "delegatorRewards" | "delegator_rewards" => Ok(GeneratedField::DelegatorRewards),
                            "delegationSequence" | "delegation_sequence" => Ok(GeneratedField::DelegationSequence),
                            "unbondingSequence" | "unbonding_sequence" => Ok(GeneratedField::UnbondingSequence),
                            "redelegationSequence" | "redelegation_sequence" => Ok(GeneratedField::RedelegationSequence),
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
                formatter.write_str("struct virtengine.delegation.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut delegations__ = None;
                let mut unbonding_delegations__ = None;
                let mut redelegations__ = None;
                let mut validator_shares__ = None;
                let mut delegator_rewards__ = None;
                let mut delegation_sequence__ = None;
                let mut unbonding_sequence__ = None;
                let mut redelegation_sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Delegations => {
                            if delegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegations"));
                            }
                            delegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UnbondingDelegations => {
                            if unbonding_delegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingDelegations"));
                            }
                            unbonding_delegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Redelegations => {
                            if redelegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("redelegations"));
                            }
                            redelegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorShares => {
                            if validator_shares__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorShares"));
                            }
                            validator_shares__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DelegatorRewards => {
                            if delegator_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorRewards"));
                            }
                            delegator_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DelegationSequence => {
                            if delegation_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegationSequence"));
                            }
                            delegation_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UnbondingSequence => {
                            if unbonding_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingSequence"));
                            }
                            unbonding_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RedelegationSequence => {
                            if redelegation_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("redelegationSequence"));
                            }
                            redelegation_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    delegations: delegations__.unwrap_or_default(),
                    unbonding_delegations: unbonding_delegations__.unwrap_or_default(),
                    redelegations: redelegations__.unwrap_or_default(),
                    validator_shares: validator_shares__.unwrap_or_default(),
                    delegator_rewards: delegator_rewards__.unwrap_or_default(),
                    delegation_sequence: delegation_sequence__.unwrap_or_default(),
                    unbonding_sequence: unbonding_sequence__.unwrap_or_default(),
                    redelegation_sequence: redelegation_sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimAllRewards {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgClaimAllRewards", len)?;
        if !self.delegator.is_empty() {
            struct_ser.serialize_field("delegator", &self.delegator)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimAllRewards {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegator,
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
                            "delegator" => Ok(GeneratedField::Delegator),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgClaimAllRewards;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgClaimAllRewards")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimAllRewards, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegator => {
                            if delegator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegator"));
                            }
                            delegator__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgClaimAllRewards {
                    delegator: delegator__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgClaimAllRewards", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimAllRewardsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.amount.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgClaimAllRewardsResponse", len)?;
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimAllRewardsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "amount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = MsgClaimAllRewardsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgClaimAllRewardsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimAllRewardsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgClaimAllRewardsResponse {
                    amount: amount__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgClaimAllRewardsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimRewards {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator.is_empty() {
            len += 1;
        }
        if !self.validator.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgClaimRewards", len)?;
        if !self.delegator.is_empty() {
            struct_ser.serialize_field("delegator", &self.delegator)?;
        }
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimRewards {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator",
            "validator",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegator,
            Validator,
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
                            "delegator" => Ok(GeneratedField::Delegator),
                            "validator" => Ok(GeneratedField::Validator),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgClaimRewards;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgClaimRewards")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimRewards, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator__ = None;
                let mut validator__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegator => {
                            if delegator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegator"));
                            }
                            delegator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgClaimRewards {
                    delegator: delegator__.unwrap_or_default(),
                    validator: validator__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgClaimRewards", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimRewardsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.amount.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgClaimRewardsResponse", len)?;
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimRewardsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "amount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = MsgClaimRewardsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgClaimRewardsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimRewardsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgClaimRewardsResponse {
                    amount: amount__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgClaimRewardsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDelegate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator.is_empty() {
            len += 1;
        }
        if !self.validator.is_empty() {
            len += 1;
        }
        if self.amount.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgDelegate", len)?;
        if !self.delegator.is_empty() {
            struct_ser.serialize_field("delegator", &self.delegator)?;
        }
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if let Some(v) = self.amount.as_ref() {
            struct_ser.serialize_field("amount", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDelegate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator",
            "validator",
            "amount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegator,
            Validator,
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
                            "delegator" => Ok(GeneratedField::Delegator),
                            "validator" => Ok(GeneratedField::Validator),
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
            type Value = MsgDelegate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgDelegate")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDelegate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator__ = None;
                let mut validator__ = None;
                let mut amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegator => {
                            if delegator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegator"));
                            }
                            delegator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgDelegate {
                    delegator: delegator__.unwrap_or_default(),
                    validator: validator__.unwrap_or_default(),
                    amount: amount__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgDelegate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDelegateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgDelegateResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDelegateResponse {
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
            type Value = MsgDelegateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgDelegateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDelegateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDelegateResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgDelegateResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRedelegate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator.is_empty() {
            len += 1;
        }
        if !self.src_validator.is_empty() {
            len += 1;
        }
        if !self.dst_validator.is_empty() {
            len += 1;
        }
        if self.amount.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgRedelegate", len)?;
        if !self.delegator.is_empty() {
            struct_ser.serialize_field("delegator", &self.delegator)?;
        }
        if !self.src_validator.is_empty() {
            struct_ser.serialize_field("srcValidator", &self.src_validator)?;
        }
        if !self.dst_validator.is_empty() {
            struct_ser.serialize_field("dstValidator", &self.dst_validator)?;
        }
        if let Some(v) = self.amount.as_ref() {
            struct_ser.serialize_field("amount", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRedelegate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator",
            "src_validator",
            "srcValidator",
            "dst_validator",
            "dstValidator",
            "amount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegator,
            SrcValidator,
            DstValidator,
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
                            "delegator" => Ok(GeneratedField::Delegator),
                            "srcValidator" | "src_validator" => Ok(GeneratedField::SrcValidator),
                            "dstValidator" | "dst_validator" => Ok(GeneratedField::DstValidator),
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
            type Value = MsgRedelegate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgRedelegate")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRedelegate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator__ = None;
                let mut src_validator__ = None;
                let mut dst_validator__ = None;
                let mut amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegator => {
                            if delegator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegator"));
                            }
                            delegator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SrcValidator => {
                            if src_validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("srcValidator"));
                            }
                            src_validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DstValidator => {
                            if dst_validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("dstValidator"));
                            }
                            dst_validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgRedelegate {
                    delegator: delegator__.unwrap_or_default(),
                    src_validator: src_validator__.unwrap_or_default(),
                    dst_validator: dst_validator__.unwrap_or_default(),
                    amount: amount__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgRedelegate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRedelegateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.completion_time != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgRedelegateResponse", len)?;
        if self.completion_time != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("completionTime", ToString::to_string(&self.completion_time).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRedelegateResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "completion_time",
            "completionTime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CompletionTime,
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
                            "completionTime" | "completion_time" => Ok(GeneratedField::CompletionTime),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRedelegateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgRedelegateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRedelegateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut completion_time__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CompletionTime => {
                            if completion_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completionTime"));
                            }
                            completion_time__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRedelegateResponse {
                    completion_time: completion_time__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgRedelegateResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUndelegate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator.is_empty() {
            len += 1;
        }
        if !self.validator.is_empty() {
            len += 1;
        }
        if self.amount.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgUndelegate", len)?;
        if !self.delegator.is_empty() {
            struct_ser.serialize_field("delegator", &self.delegator)?;
        }
        if !self.validator.is_empty() {
            struct_ser.serialize_field("validator", &self.validator)?;
        }
        if let Some(v) = self.amount.as_ref() {
            struct_ser.serialize_field("amount", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUndelegate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator",
            "validator",
            "amount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegator,
            Validator,
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
                            "delegator" => Ok(GeneratedField::Delegator),
                            "validator" => Ok(GeneratedField::Validator),
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
            type Value = MsgUndelegate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgUndelegate")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUndelegate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator__ = None;
                let mut validator__ = None;
                let mut amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegator => {
                            if delegator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegator"));
                            }
                            delegator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Validator => {
                            if validator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validator"));
                            }
                            validator__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgUndelegate {
                    delegator: delegator__.unwrap_or_default(),
                    validator: validator__.unwrap_or_default(),
                    amount: amount__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgUndelegate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUndelegateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.completion_time != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgUndelegateResponse", len)?;
        if self.completion_time != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("completionTime", ToString::to_string(&self.completion_time).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUndelegateResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "completion_time",
            "completionTime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CompletionTime,
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
                            "completionTime" | "completion_time" => Ok(GeneratedField::CompletionTime),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUndelegateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.MsgUndelegateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUndelegateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut completion_time__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CompletionTime => {
                            if completion_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completionTime"));
                            }
                            completion_time__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUndelegateResponse {
                    completion_time: completion_time__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgUndelegateResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.delegation.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.delegation.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.delegation.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.delegation.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if self.unbonding_period != 0 {
            len += 1;
        }
        if self.max_validators_per_delegator != 0 {
            len += 1;
        }
        if self.min_delegation_amount != 0 {
            len += 1;
        }
        if self.max_redelegations != 0 {
            len += 1;
        }
        if self.validator_commission_rate != 0 {
            len += 1;
        }
        if !self.reward_denom.is_empty() {
            len += 1;
        }
        if !self.stake_denom.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.Params", len)?;
        if self.unbonding_period != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("unbondingPeriod", ToString::to_string(&self.unbonding_period).as_str())?;
        }
        if self.max_validators_per_delegator != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxValidatorsPerDelegator", ToString::to_string(&self.max_validators_per_delegator).as_str())?;
        }
        if self.min_delegation_amount != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("minDelegationAmount", ToString::to_string(&self.min_delegation_amount).as_str())?;
        }
        if self.max_redelegations != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxRedelegations", ToString::to_string(&self.max_redelegations).as_str())?;
        }
        if self.validator_commission_rate != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("validatorCommissionRate", ToString::to_string(&self.validator_commission_rate).as_str())?;
        }
        if !self.reward_denom.is_empty() {
            struct_ser.serialize_field("rewardDenom", &self.reward_denom)?;
        }
        if !self.stake_denom.is_empty() {
            struct_ser.serialize_field("stakeDenom", &self.stake_denom)?;
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
            "unbonding_period",
            "unbondingPeriod",
            "max_validators_per_delegator",
            "maxValidatorsPerDelegator",
            "min_delegation_amount",
            "minDelegationAmount",
            "max_redelegations",
            "maxRedelegations",
            "validator_commission_rate",
            "validatorCommissionRate",
            "reward_denom",
            "rewardDenom",
            "stake_denom",
            "stakeDenom",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            UnbondingPeriod,
            MaxValidatorsPerDelegator,
            MinDelegationAmount,
            MaxRedelegations,
            ValidatorCommissionRate,
            RewardDenom,
            StakeDenom,
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
                            "unbondingPeriod" | "unbonding_period" => Ok(GeneratedField::UnbondingPeriod),
                            "maxValidatorsPerDelegator" | "max_validators_per_delegator" => Ok(GeneratedField::MaxValidatorsPerDelegator),
                            "minDelegationAmount" | "min_delegation_amount" => Ok(GeneratedField::MinDelegationAmount),
                            "maxRedelegations" | "max_redelegations" => Ok(GeneratedField::MaxRedelegations),
                            "validatorCommissionRate" | "validator_commission_rate" => Ok(GeneratedField::ValidatorCommissionRate),
                            "rewardDenom" | "reward_denom" => Ok(GeneratedField::RewardDenom),
                            "stakeDenom" | "stake_denom" => Ok(GeneratedField::StakeDenom),
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
                formatter.write_str("struct virtengine.delegation.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut unbonding_period__ = None;
                let mut max_validators_per_delegator__ = None;
                let mut min_delegation_amount__ = None;
                let mut max_redelegations__ = None;
                let mut validator_commission_rate__ = None;
                let mut reward_denom__ = None;
                let mut stake_denom__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UnbondingPeriod => {
                            if unbonding_period__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingPeriod"));
                            }
                            unbonding_period__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxValidatorsPerDelegator => {
                            if max_validators_per_delegator__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxValidatorsPerDelegator"));
                            }
                            max_validators_per_delegator__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinDelegationAmount => {
                            if min_delegation_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minDelegationAmount"));
                            }
                            min_delegation_amount__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRedelegations => {
                            if max_redelegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRedelegations"));
                            }
                            max_redelegations__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ValidatorCommissionRate => {
                            if validator_commission_rate__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorCommissionRate"));
                            }
                            validator_commission_rate__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RewardDenom => {
                            if reward_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardDenom"));
                            }
                            reward_denom__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StakeDenom => {
                            if stake_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("stakeDenom"));
                            }
                            stake_denom__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Params {
                    unbonding_period: unbonding_period__.unwrap_or_default(),
                    max_validators_per_delegator: max_validators_per_delegator__.unwrap_or_default(),
                    min_delegation_amount: min_delegation_amount__.unwrap_or_default(),
                    max_redelegations: max_redelegations__.unwrap_or_default(),
                    validator_commission_rate: validator_commission_rate__.unwrap_or_default(),
                    reward_denom: reward_denom__.unwrap_or_default(),
                    stake_denom: stake_denom__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegationRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegationRequest", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegationRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "validator_address",
            "validatorAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
            ValidatorAddress,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDelegationRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegationRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegationRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut validator_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDelegationRequest {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegationRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.delegation.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegationResponse", len)?;
        if let Some(v) = self.delegation.as_ref() {
            struct_ser.serialize_field("delegation", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegationResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegation",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegation,
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
                            "delegation" => Ok(GeneratedField::Delegation),
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
            type Value = QueryDelegationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegation__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegation => {
                            if delegation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegation"));
                            }
                            delegation__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDelegationResponse {
                    delegation: delegation__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorAllRewardsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorAllRewardsRequest", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorAllRewardsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
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
            type Value = QueryDelegatorAllRewardsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorAllRewardsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorAllRewardsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorAllRewardsRequest {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorAllRewardsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorAllRewardsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.rewards.is_empty() {
            len += 1;
        }
        if !self.total_reward.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorAllRewardsResponse", len)?;
        if !self.rewards.is_empty() {
            struct_ser.serialize_field("rewards", &self.rewards)?;
        }
        if !self.total_reward.is_empty() {
            struct_ser.serialize_field("totalReward", &self.total_reward)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorAllRewardsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "rewards",
            "total_reward",
            "totalReward",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Rewards,
            TotalReward,
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
                            "rewards" => Ok(GeneratedField::Rewards),
                            "totalReward" | "total_reward" => Ok(GeneratedField::TotalReward),
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
            type Value = QueryDelegatorAllRewardsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorAllRewardsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorAllRewardsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut rewards__ = None;
                let mut total_reward__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Rewards => {
                            if rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewards"));
                            }
                            rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalReward => {
                            if total_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalReward"));
                            }
                            total_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorAllRewardsResponse {
                    rewards: rewards__.unwrap_or_default(),
                    total_reward: total_reward__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorAllRewardsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorDelegationsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorDelegationsRequest", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorDelegationsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
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
            type Value = QueryDelegatorDelegationsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorDelegationsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorDelegationsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorDelegationsRequest {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorDelegationsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorDelegationsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegations.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorDelegationsResponse", len)?;
        if !self.delegations.is_empty() {
            struct_ser.serialize_field("delegations", &self.delegations)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorDelegationsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegations",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegations,
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
                            "delegations" => Ok(GeneratedField::Delegations),
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
            type Value = QueryDelegatorDelegationsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorDelegationsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorDelegationsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegations__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegations => {
                            if delegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegations"));
                            }
                            delegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorDelegationsResponse {
                    delegations: delegations__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorDelegationsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorRedelegationsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorRedelegationsRequest", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorRedelegationsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
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
            type Value = QueryDelegatorRedelegationsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorRedelegationsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorRedelegationsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorRedelegationsRequest {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorRedelegationsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorRedelegationsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.redelegations.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorRedelegationsResponse", len)?;
        if !self.redelegations.is_empty() {
            struct_ser.serialize_field("redelegations", &self.redelegations)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorRedelegationsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "redelegations",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Redelegations,
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
                            "redelegations" => Ok(GeneratedField::Redelegations),
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
            type Value = QueryDelegatorRedelegationsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorRedelegationsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorRedelegationsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut redelegations__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Redelegations => {
                            if redelegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("redelegations"));
                            }
                            redelegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorRedelegationsResponse {
                    redelegations: redelegations__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorRedelegationsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorRewardsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorRewardsRequest", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorRewardsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "validator_address",
            "validatorAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
            ValidatorAddress,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDelegatorRewardsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorRewardsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorRewardsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut validator_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDelegatorRewardsRequest {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorRewardsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorRewardsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.rewards.is_empty() {
            len += 1;
        }
        if !self.total_reward.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorRewardsResponse", len)?;
        if !self.rewards.is_empty() {
            struct_ser.serialize_field("rewards", &self.rewards)?;
        }
        if !self.total_reward.is_empty() {
            struct_ser.serialize_field("totalReward", &self.total_reward)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorRewardsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "rewards",
            "total_reward",
            "totalReward",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Rewards,
            TotalReward,
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
                            "rewards" => Ok(GeneratedField::Rewards),
                            "totalReward" | "total_reward" => Ok(GeneratedField::TotalReward),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDelegatorRewardsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorRewardsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorRewardsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut rewards__ = None;
                let mut total_reward__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Rewards => {
                            if rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewards"));
                            }
                            rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalReward => {
                            if total_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalReward"));
                            }
                            total_reward__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDelegatorRewardsResponse {
                    rewards: rewards__.unwrap_or_default(),
                    total_reward: total_reward__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorRewardsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorUnbondingDelegationsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest", len)?;
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorUnbondingDelegationsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegator_address",
            "delegatorAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelegatorAddress,
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
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
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
            type Value = QueryDelegatorUnbondingDelegationsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorUnbondingDelegationsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegator_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorUnbondingDelegationsRequest {
                    delegator_address: delegator_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDelegatorUnbondingDelegationsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.unbonding_delegations.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse", len)?;
        if !self.unbonding_delegations.is_empty() {
            struct_ser.serialize_field("unbondingDelegations", &self.unbonding_delegations)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDelegatorUnbondingDelegationsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "unbonding_delegations",
            "unbondingDelegations",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            UnbondingDelegations,
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
                            "unbondingDelegations" | "unbonding_delegations" => Ok(GeneratedField::UnbondingDelegations),
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
            type Value = QueryDelegatorUnbondingDelegationsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDelegatorUnbondingDelegationsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut unbonding_delegations__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UnbondingDelegations => {
                            if unbonding_delegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingDelegations"));
                            }
                            unbonding_delegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryDelegatorUnbondingDelegationsResponse {
                    unbonding_delegations: unbonding_delegations__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryDelegatorUnbondingDelegationsResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.delegation.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.delegation.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRedelegationRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.redelegation_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryRedelegationRequest", len)?;
        if !self.redelegation_id.is_empty() {
            struct_ser.serialize_field("redelegationId", &self.redelegation_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRedelegationRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "redelegation_id",
            "redelegationId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RedelegationId,
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
                            "redelegationId" | "redelegation_id" => Ok(GeneratedField::RedelegationId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryRedelegationRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryRedelegationRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRedelegationRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut redelegation_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RedelegationId => {
                            if redelegation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("redelegationId"));
                            }
                            redelegation_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryRedelegationRequest {
                    redelegation_id: redelegation_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryRedelegationRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryRedelegationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.redelegation.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryRedelegationResponse", len)?;
        if let Some(v) = self.redelegation.as_ref() {
            struct_ser.serialize_field("redelegation", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryRedelegationResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "redelegation",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Redelegation,
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
                            "redelegation" => Ok(GeneratedField::Redelegation),
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
            type Value = QueryRedelegationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryRedelegationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryRedelegationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut redelegation__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Redelegation => {
                            if redelegation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("redelegation"));
                            }
                            redelegation__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryRedelegationResponse {
                    redelegation: redelegation__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryRedelegationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryUnbondingDelegationRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.unbonding_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryUnbondingDelegationRequest", len)?;
        if !self.unbonding_id.is_empty() {
            struct_ser.serialize_field("unbondingId", &self.unbonding_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryUnbondingDelegationRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "unbonding_id",
            "unbondingId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            UnbondingId,
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
                            "unbondingId" | "unbonding_id" => Ok(GeneratedField::UnbondingId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryUnbondingDelegationRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryUnbondingDelegationRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryUnbondingDelegationRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut unbonding_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UnbondingId => {
                            if unbonding_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingId"));
                            }
                            unbonding_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryUnbondingDelegationRequest {
                    unbonding_id: unbonding_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryUnbondingDelegationRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryUnbondingDelegationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.unbonding_delegation.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryUnbondingDelegationResponse", len)?;
        if let Some(v) = self.unbonding_delegation.as_ref() {
            struct_ser.serialize_field("unbondingDelegation", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryUnbondingDelegationResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "unbonding_delegation",
            "unbondingDelegation",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            UnbondingDelegation,
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
                            "unbondingDelegation" | "unbonding_delegation" => Ok(GeneratedField::UnbondingDelegation),
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
            type Value = QueryUnbondingDelegationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryUnbondingDelegationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryUnbondingDelegationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut unbonding_delegation__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UnbondingDelegation => {
                            if unbonding_delegation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingDelegation"));
                            }
                            unbonding_delegation__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryUnbondingDelegationResponse {
                    unbonding_delegation: unbonding_delegation__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryUnbondingDelegationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidatorDelegationsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryValidatorDelegationsRequest", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidatorDelegationsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
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
            type Value = QueryValidatorDelegationsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryValidatorDelegationsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidatorDelegationsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryValidatorDelegationsRequest {
                    validator_address: validator_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryValidatorDelegationsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidatorDelegationsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.delegations.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryValidatorDelegationsResponse", len)?;
        if !self.delegations.is_empty() {
            struct_ser.serialize_field("delegations", &self.delegations)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidatorDelegationsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "delegations",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Delegations,
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
                            "delegations" => Ok(GeneratedField::Delegations),
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
            type Value = QueryValidatorDelegationsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryValidatorDelegationsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidatorDelegationsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut delegations__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Delegations => {
                            if delegations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegations"));
                            }
                            delegations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryValidatorDelegationsResponse {
                    delegations: delegations__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryValidatorDelegationsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidatorSharesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryValidatorSharesRequest", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidatorSharesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryValidatorSharesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryValidatorSharesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidatorSharesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryValidatorSharesRequest {
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryValidatorSharesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidatorSharesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.validator_shares.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.QueryValidatorSharesResponse", len)?;
        if let Some(v) = self.validator_shares.as_ref() {
            struct_ser.serialize_field("validatorShares", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidatorSharesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_shares",
            "validatorShares",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorShares,
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
                            "validatorShares" | "validator_shares" => Ok(GeneratedField::ValidatorShares),
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
            type Value = QueryValidatorSharesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.QueryValidatorSharesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidatorSharesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_shares__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorShares => {
                            if validator_shares__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorShares"));
                            }
                            validator_shares__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryValidatorSharesResponse {
                    validator_shares: validator_shares__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.QueryValidatorSharesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Redelegation {
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
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if !self.validator_src_address.is_empty() {
            len += 1;
        }
        if !self.validator_dst_address.is_empty() {
            len += 1;
        }
        if !self.entries.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.Redelegation", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if !self.validator_src_address.is_empty() {
            struct_ser.serialize_field("validatorSrcAddress", &self.validator_src_address)?;
        }
        if !self.validator_dst_address.is_empty() {
            struct_ser.serialize_field("validatorDstAddress", &self.validator_dst_address)?;
        }
        if !self.entries.is_empty() {
            struct_ser.serialize_field("entries", &self.entries)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Redelegation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "delegator_address",
            "delegatorAddress",
            "validator_src_address",
            "validatorSrcAddress",
            "validator_dst_address",
            "validatorDstAddress",
            "entries",
            "created_at",
            "createdAt",
            "height",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            DelegatorAddress,
            ValidatorSrcAddress,
            ValidatorDstAddress,
            Entries,
            CreatedAt,
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
                            "id" => Ok(GeneratedField::Id),
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
                            "validatorSrcAddress" | "validator_src_address" => Ok(GeneratedField::ValidatorSrcAddress),
                            "validatorDstAddress" | "validator_dst_address" => Ok(GeneratedField::ValidatorDstAddress),
                            "entries" => Ok(GeneratedField::Entries),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
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
            type Value = Redelegation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.Redelegation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Redelegation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut delegator_address__ = None;
                let mut validator_src_address__ = None;
                let mut validator_dst_address__ = None;
                let mut entries__ = None;
                let mut created_at__ = None;
                let mut height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorSrcAddress => {
                            if validator_src_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorSrcAddress"));
                            }
                            validator_src_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorDstAddress => {
                            if validator_dst_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorDstAddress"));
                            }
                            validator_dst_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Entries => {
                            if entries__.is_some() {
                                return Err(serde::de::Error::duplicate_field("entries"));
                            }
                            entries__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
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
                Ok(Redelegation {
                    id: id__.unwrap_or_default(),
                    delegator_address: delegator_address__.unwrap_or_default(),
                    validator_src_address: validator_src_address__.unwrap_or_default(),
                    validator_dst_address: validator_dst_address__.unwrap_or_default(),
                    entries: entries__.unwrap_or_default(),
                    created_at: created_at__,
                    height: height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.Redelegation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RedelegationEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.creation_height != 0 {
            len += 1;
        }
        if self.completion_time.is_some() {
            len += 1;
        }
        if !self.initial_balance.is_empty() {
            len += 1;
        }
        if !self.shares_dst.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.RedelegationEntry", len)?;
        if self.creation_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("creationHeight", ToString::to_string(&self.creation_height).as_str())?;
        }
        if let Some(v) = self.completion_time.as_ref() {
            struct_ser.serialize_field("completionTime", v)?;
        }
        if !self.initial_balance.is_empty() {
            struct_ser.serialize_field("initialBalance", &self.initial_balance)?;
        }
        if !self.shares_dst.is_empty() {
            struct_ser.serialize_field("sharesDst", &self.shares_dst)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RedelegationEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "creation_height",
            "creationHeight",
            "completion_time",
            "completionTime",
            "initial_balance",
            "initialBalance",
            "shares_dst",
            "sharesDst",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CreationHeight,
            CompletionTime,
            InitialBalance,
            SharesDst,
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
                            "creationHeight" | "creation_height" => Ok(GeneratedField::CreationHeight),
                            "completionTime" | "completion_time" => Ok(GeneratedField::CompletionTime),
                            "initialBalance" | "initial_balance" => Ok(GeneratedField::InitialBalance),
                            "sharesDst" | "shares_dst" => Ok(GeneratedField::SharesDst),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RedelegationEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.RedelegationEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RedelegationEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut creation_height__ = None;
                let mut completion_time__ = None;
                let mut initial_balance__ = None;
                let mut shares_dst__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CreationHeight => {
                            if creation_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("creationHeight"));
                            }
                            creation_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CompletionTime => {
                            if completion_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completionTime"));
                            }
                            completion_time__ = map_.next_value()?;
                        }
                        GeneratedField::InitialBalance => {
                            if initial_balance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("initialBalance"));
                            }
                            initial_balance__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SharesDst => {
                            if shares_dst__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sharesDst"));
                            }
                            shares_dst__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(RedelegationEntry {
                    creation_height: creation_height__.unwrap_or_default(),
                    completion_time: completion_time__,
                    initial_balance: initial_balance__.unwrap_or_default(),
                    shares_dst: shares_dst__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.RedelegationEntry", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for UnbondingDelegation {
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
        if !self.delegator_address.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.entries.is_empty() {
            len += 1;
        }
        if self.created_at.is_some() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.UnbondingDelegation", len)?;
        if !self.id.is_empty() {
            struct_ser.serialize_field("id", &self.id)?;
        }
        if !self.delegator_address.is_empty() {
            struct_ser.serialize_field("delegatorAddress", &self.delegator_address)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.entries.is_empty() {
            struct_ser.serialize_field("entries", &self.entries)?;
        }
        if let Some(v) = self.created_at.as_ref() {
            struct_ser.serialize_field("createdAt", v)?;
        }
        if self.height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for UnbondingDelegation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "id",
            "delegator_address",
            "delegatorAddress",
            "validator_address",
            "validatorAddress",
            "entries",
            "created_at",
            "createdAt",
            "height",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Id,
            DelegatorAddress,
            ValidatorAddress,
            Entries,
            CreatedAt,
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
                            "id" => Ok(GeneratedField::Id),
                            "delegatorAddress" | "delegator_address" => Ok(GeneratedField::DelegatorAddress),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "entries" => Ok(GeneratedField::Entries),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
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
            type Value = UnbondingDelegation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.UnbondingDelegation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<UnbondingDelegation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut id__ = None;
                let mut delegator_address__ = None;
                let mut validator_address__ = None;
                let mut entries__ = None;
                let mut created_at__ = None;
                let mut height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Id => {
                            if id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("id"));
                            }
                            id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DelegatorAddress => {
                            if delegator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delegatorAddress"));
                            }
                            delegator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Entries => {
                            if entries__.is_some() {
                                return Err(serde::de::Error::duplicate_field("entries"));
                            }
                            entries__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = map_.next_value()?;
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
                Ok(UnbondingDelegation {
                    id: id__.unwrap_or_default(),
                    delegator_address: delegator_address__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    entries: entries__.unwrap_or_default(),
                    created_at: created_at__,
                    height: height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.UnbondingDelegation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for UnbondingDelegationEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.creation_height != 0 {
            len += 1;
        }
        if self.completion_time.is_some() {
            len += 1;
        }
        if !self.initial_balance.is_empty() {
            len += 1;
        }
        if !self.balance.is_empty() {
            len += 1;
        }
        if !self.unbonding_shares.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.UnbondingDelegationEntry", len)?;
        if self.creation_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("creationHeight", ToString::to_string(&self.creation_height).as_str())?;
        }
        if let Some(v) = self.completion_time.as_ref() {
            struct_ser.serialize_field("completionTime", v)?;
        }
        if !self.initial_balance.is_empty() {
            struct_ser.serialize_field("initialBalance", &self.initial_balance)?;
        }
        if !self.balance.is_empty() {
            struct_ser.serialize_field("balance", &self.balance)?;
        }
        if !self.unbonding_shares.is_empty() {
            struct_ser.serialize_field("unbondingShares", &self.unbonding_shares)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for UnbondingDelegationEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "creation_height",
            "creationHeight",
            "completion_time",
            "completionTime",
            "initial_balance",
            "initialBalance",
            "balance",
            "unbonding_shares",
            "unbondingShares",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CreationHeight,
            CompletionTime,
            InitialBalance,
            Balance,
            UnbondingShares,
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
                            "creationHeight" | "creation_height" => Ok(GeneratedField::CreationHeight),
                            "completionTime" | "completion_time" => Ok(GeneratedField::CompletionTime),
                            "initialBalance" | "initial_balance" => Ok(GeneratedField::InitialBalance),
                            "balance" => Ok(GeneratedField::Balance),
                            "unbondingShares" | "unbonding_shares" => Ok(GeneratedField::UnbondingShares),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = UnbondingDelegationEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.UnbondingDelegationEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<UnbondingDelegationEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut creation_height__ = None;
                let mut completion_time__ = None;
                let mut initial_balance__ = None;
                let mut balance__ = None;
                let mut unbonding_shares__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CreationHeight => {
                            if creation_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("creationHeight"));
                            }
                            creation_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CompletionTime => {
                            if completion_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("completionTime"));
                            }
                            completion_time__ = map_.next_value()?;
                        }
                        GeneratedField::InitialBalance => {
                            if initial_balance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("initialBalance"));
                            }
                            initial_balance__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Balance => {
                            if balance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("balance"));
                            }
                            balance__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UnbondingShares => {
                            if unbonding_shares__.is_some() {
                                return Err(serde::de::Error::duplicate_field("unbondingShares"));
                            }
                            unbonding_shares__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(UnbondingDelegationEntry {
                    creation_height: creation_height__.unwrap_or_default(),
                    completion_time: completion_time__,
                    initial_balance: initial_balance__.unwrap_or_default(),
                    balance: balance__.unwrap_or_default(),
                    unbonding_shares: unbonding_shares__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.UnbondingDelegationEntry", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorShares {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.total_shares.is_empty() {
            len += 1;
        }
        if !self.total_stake.is_empty() {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.delegation.v1.ValidatorShares", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.total_shares.is_empty() {
            struct_ser.serialize_field("totalShares", &self.total_shares)?;
        }
        if !self.total_stake.is_empty() {
            struct_ser.serialize_field("totalStake", &self.total_stake)?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValidatorShares {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "total_shares",
            "totalShares",
            "total_stake",
            "totalStake",
            "updated_at",
            "updatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            TotalShares,
            TotalStake,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "totalShares" | "total_shares" => Ok(GeneratedField::TotalShares),
                            "totalStake" | "total_stake" => Ok(GeneratedField::TotalStake),
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
            type Value = ValidatorShares;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.delegation.v1.ValidatorShares")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValidatorShares, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut total_shares__ = None;
                let mut total_stake__ = None;
                let mut updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalShares => {
                            if total_shares__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalShares"));
                            }
                            total_shares__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalStake => {
                            if total_stake__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalStake"));
                            }
                            total_stake__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ValidatorShares {
                    validator_address: validator_address__.unwrap_or_default(),
                    total_shares: total_shares__.unwrap_or_default(),
                    total_stake: total_stake__.unwrap_or_default(),
                    updated_at: updated_at__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.delegation.v1.ValidatorShares", FIELDS, GeneratedVisitor)
    }
}

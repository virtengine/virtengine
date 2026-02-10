// @generated
impl serde::Serialize for DoubleSignEvidence {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.evidence_id.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.height_1 != 0 {
            len += 1;
        }
        if self.height_2 != 0 {
            len += 1;
        }
        if !self.vote_hash_1.is_empty() {
            len += 1;
        }
        if !self.vote_hash_2.is_empty() {
            len += 1;
        }
        if self.detected_at.is_some() {
            len += 1;
        }
        if self.detected_height != 0 {
            len += 1;
        }
        if self.processed {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.DoubleSignEvidence", len)?;
        if !self.evidence_id.is_empty() {
            struct_ser.serialize_field("evidenceId", &self.evidence_id)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.height_1 != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height1", ToString::to_string(&self.height_1).as_str())?;
        }
        if self.height_2 != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("height2", ToString::to_string(&self.height_2).as_str())?;
        }
        if !self.vote_hash_1.is_empty() {
            struct_ser.serialize_field("voteHash1", &self.vote_hash_1)?;
        }
        if !self.vote_hash_2.is_empty() {
            struct_ser.serialize_field("voteHash2", &self.vote_hash_2)?;
        }
        if let Some(v) = self.detected_at.as_ref() {
            struct_ser.serialize_field("detectedAt", v)?;
        }
        if self.detected_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("detectedHeight", ToString::to_string(&self.detected_height).as_str())?;
        }
        if self.processed {
            struct_ser.serialize_field("processed", &self.processed)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DoubleSignEvidence {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "evidence_id",
            "evidenceId",
            "validator_address",
            "validatorAddress",
            "height_1",
            "height1",
            "height_2",
            "height2",
            "vote_hash_1",
            "voteHash1",
            "vote_hash_2",
            "voteHash2",
            "detected_at",
            "detectedAt",
            "detected_height",
            "detectedHeight",
            "processed",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EvidenceId,
            ValidatorAddress,
            Height1,
            Height2,
            VoteHash1,
            VoteHash2,
            DetectedAt,
            DetectedHeight,
            Processed,
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
                            "evidenceId" | "evidence_id" => Ok(GeneratedField::EvidenceId),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "height1" | "height_1" => Ok(GeneratedField::Height1),
                            "height2" | "height_2" => Ok(GeneratedField::Height2),
                            "voteHash1" | "vote_hash_1" => Ok(GeneratedField::VoteHash1),
                            "voteHash2" | "vote_hash_2" => Ok(GeneratedField::VoteHash2),
                            "detectedAt" | "detected_at" => Ok(GeneratedField::DetectedAt),
                            "detectedHeight" | "detected_height" => Ok(GeneratedField::DetectedHeight),
                            "processed" => Ok(GeneratedField::Processed),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DoubleSignEvidence;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.DoubleSignEvidence")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DoubleSignEvidence, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut evidence_id__ = None;
                let mut validator_address__ = None;
                let mut height_1__ = None;
                let mut height_2__ = None;
                let mut vote_hash_1__ = None;
                let mut vote_hash_2__ = None;
                let mut detected_at__ = None;
                let mut detected_height__ = None;
                let mut processed__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EvidenceId => {
                            if evidence_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidenceId"));
                            }
                            evidence_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Height1 => {
                            if height_1__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height1"));
                            }
                            height_1__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Height2 => {
                            if height_2__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height2"));
                            }
                            height_2__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VoteHash1 => {
                            if vote_hash_1__.is_some() {
                                return Err(serde::de::Error::duplicate_field("voteHash1"));
                            }
                            vote_hash_1__ = Some(map_.next_value()?);
                        }
                        GeneratedField::VoteHash2 => {
                            if vote_hash_2__.is_some() {
                                return Err(serde::de::Error::duplicate_field("voteHash2"));
                            }
                            vote_hash_2__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DetectedAt => {
                            if detected_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("detectedAt"));
                            }
                            detected_at__ = map_.next_value()?;
                        }
                        GeneratedField::DetectedHeight => {
                            if detected_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("detectedHeight"));
                            }
                            detected_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Processed => {
                            if processed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("processed"));
                            }
                            processed__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(DoubleSignEvidence {
                    evidence_id: evidence_id__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    height_1: height_1__.unwrap_or_default(),
                    height_2: height_2__.unwrap_or_default(),
                    vote_hash_1: vote_hash_1__.unwrap_or_default(),
                    vote_hash_2: vote_hash_2__.unwrap_or_default(),
                    detected_at: detected_at__,
                    detected_height: detected_height__.unwrap_or_default(),
                    processed: processed__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.DoubleSignEvidence", FIELDS, GeneratedVisitor)
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
        if !self.validator_performances.is_empty() {
            len += 1;
        }
        if !self.slash_records.is_empty() {
            len += 1;
        }
        if !self.reward_epochs.is_empty() {
            len += 1;
        }
        if !self.validator_rewards.is_empty() {
            len += 1;
        }
        if !self.validator_signing_infos.is_empty() {
            len += 1;
        }
        if self.current_epoch != 0 {
            len += 1;
        }
        if self.slash_sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.validator_performances.is_empty() {
            struct_ser.serialize_field("validatorPerformances", &self.validator_performances)?;
        }
        if !self.slash_records.is_empty() {
            struct_ser.serialize_field("slashRecords", &self.slash_records)?;
        }
        if !self.reward_epochs.is_empty() {
            struct_ser.serialize_field("rewardEpochs", &self.reward_epochs)?;
        }
        if !self.validator_rewards.is_empty() {
            struct_ser.serialize_field("validatorRewards", &self.validator_rewards)?;
        }
        if !self.validator_signing_infos.is_empty() {
            struct_ser.serialize_field("validatorSigningInfos", &self.validator_signing_infos)?;
        }
        if self.current_epoch != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("currentEpoch", ToString::to_string(&self.current_epoch).as_str())?;
        }
        if self.slash_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashSequence", ToString::to_string(&self.slash_sequence).as_str())?;
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
            "validator_performances",
            "validatorPerformances",
            "slash_records",
            "slashRecords",
            "reward_epochs",
            "rewardEpochs",
            "validator_rewards",
            "validatorRewards",
            "validator_signing_infos",
            "validatorSigningInfos",
            "current_epoch",
            "currentEpoch",
            "slash_sequence",
            "slashSequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            ValidatorPerformances,
            SlashRecords,
            RewardEpochs,
            ValidatorRewards,
            ValidatorSigningInfos,
            CurrentEpoch,
            SlashSequence,
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
                            "validatorPerformances" | "validator_performances" => Ok(GeneratedField::ValidatorPerformances),
                            "slashRecords" | "slash_records" => Ok(GeneratedField::SlashRecords),
                            "rewardEpochs" | "reward_epochs" => Ok(GeneratedField::RewardEpochs),
                            "validatorRewards" | "validator_rewards" => Ok(GeneratedField::ValidatorRewards),
                            "validatorSigningInfos" | "validator_signing_infos" => Ok(GeneratedField::ValidatorSigningInfos),
                            "currentEpoch" | "current_epoch" => Ok(GeneratedField::CurrentEpoch),
                            "slashSequence" | "slash_sequence" => Ok(GeneratedField::SlashSequence),
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
                formatter.write_str("struct virtengine.staking.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut validator_performances__ = None;
                let mut slash_records__ = None;
                let mut reward_epochs__ = None;
                let mut validator_rewards__ = None;
                let mut validator_signing_infos__ = None;
                let mut current_epoch__ = None;
                let mut slash_sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::ValidatorPerformances => {
                            if validator_performances__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorPerformances"));
                            }
                            validator_performances__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SlashRecords => {
                            if slash_records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashRecords"));
                            }
                            slash_records__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RewardEpochs => {
                            if reward_epochs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardEpochs"));
                            }
                            reward_epochs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorRewards => {
                            if validator_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorRewards"));
                            }
                            validator_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorSigningInfos => {
                            if validator_signing_infos__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorSigningInfos"));
                            }
                            validator_signing_infos__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CurrentEpoch => {
                            if current_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentEpoch"));
                            }
                            current_epoch__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SlashSequence => {
                            if slash_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashSequence"));
                            }
                            slash_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    validator_performances: validator_performances__.unwrap_or_default(),
                    slash_records: slash_records__.unwrap_or_default(),
                    reward_epochs: reward_epochs__.unwrap_or_default(),
                    validator_rewards: validator_rewards__.unwrap_or_default(),
                    validator_signing_infos: validator_signing_infos__.unwrap_or_default(),
                    current_epoch: current_epoch__.unwrap_or_default(),
                    slash_sequence: slash_sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for InvalidVeidAttestation {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.record_id.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.attestation_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if self.expected_score != 0 {
            len += 1;
        }
        if self.actual_score != 0 {
            len += 1;
        }
        if self.score_difference != 0 {
            len += 1;
        }
        if self.detected_at.is_some() {
            len += 1;
        }
        if self.detected_height != 0 {
            len += 1;
        }
        if self.processed {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.InvalidVEIDAttestation", len)?;
        if !self.record_id.is_empty() {
            struct_ser.serialize_field("recordId", &self.record_id)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.attestation_id.is_empty() {
            struct_ser.serialize_field("attestationId", &self.attestation_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if self.expected_score != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expectedScore", ToString::to_string(&self.expected_score).as_str())?;
        }
        if self.actual_score != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("actualScore", ToString::to_string(&self.actual_score).as_str())?;
        }
        if self.score_difference != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("scoreDifference", ToString::to_string(&self.score_difference).as_str())?;
        }
        if let Some(v) = self.detected_at.as_ref() {
            struct_ser.serialize_field("detectedAt", v)?;
        }
        if self.detected_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("detectedHeight", ToString::to_string(&self.detected_height).as_str())?;
        }
        if self.processed {
            struct_ser.serialize_field("processed", &self.processed)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for InvalidVeidAttestation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "record_id",
            "recordId",
            "validator_address",
            "validatorAddress",
            "attestation_id",
            "attestationId",
            "reason",
            "expected_score",
            "expectedScore",
            "actual_score",
            "actualScore",
            "score_difference",
            "scoreDifference",
            "detected_at",
            "detectedAt",
            "detected_height",
            "detectedHeight",
            "processed",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RecordId,
            ValidatorAddress,
            AttestationId,
            Reason,
            ExpectedScore,
            ActualScore,
            ScoreDifference,
            DetectedAt,
            DetectedHeight,
            Processed,
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
                            "recordId" | "record_id" => Ok(GeneratedField::RecordId),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "attestationId" | "attestation_id" => Ok(GeneratedField::AttestationId),
                            "reason" => Ok(GeneratedField::Reason),
                            "expectedScore" | "expected_score" => Ok(GeneratedField::ExpectedScore),
                            "actualScore" | "actual_score" => Ok(GeneratedField::ActualScore),
                            "scoreDifference" | "score_difference" => Ok(GeneratedField::ScoreDifference),
                            "detectedAt" | "detected_at" => Ok(GeneratedField::DetectedAt),
                            "detectedHeight" | "detected_height" => Ok(GeneratedField::DetectedHeight),
                            "processed" => Ok(GeneratedField::Processed),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = InvalidVeidAttestation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.InvalidVEIDAttestation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<InvalidVeidAttestation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut record_id__ = None;
                let mut validator_address__ = None;
                let mut attestation_id__ = None;
                let mut reason__ = None;
                let mut expected_score__ = None;
                let mut actual_score__ = None;
                let mut score_difference__ = None;
                let mut detected_at__ = None;
                let mut detected_height__ = None;
                let mut processed__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RecordId => {
                            if record_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recordId"));
                            }
                            record_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AttestationId => {
                            if attestation_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationId"));
                            }
                            attestation_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpectedScore => {
                            if expected_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expectedScore"));
                            }
                            expected_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ActualScore => {
                            if actual_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("actualScore"));
                            }
                            actual_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScoreDifference => {
                            if score_difference__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreDifference"));
                            }
                            score_difference__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DetectedAt => {
                            if detected_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("detectedAt"));
                            }
                            detected_at__ = map_.next_value()?;
                        }
                        GeneratedField::DetectedHeight => {
                            if detected_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("detectedHeight"));
                            }
                            detected_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Processed => {
                            if processed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("processed"));
                            }
                            processed__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(InvalidVeidAttestation {
                    record_id: record_id__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    attestation_id: attestation_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    expected_score: expected_score__.unwrap_or_default(),
                    actual_score: actual_score__.unwrap_or_default(),
                    score_difference: score_difference__.unwrap_or_default(),
                    detected_at: detected_at__,
                    detected_height: detected_height__.unwrap_or_default(),
                    processed: processed__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.InvalidVEIDAttestation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRecordPerformance {
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
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.blocks_proposed != 0 {
            len += 1;
        }
        if self.blocks_signed != 0 {
            len += 1;
        }
        if self.veid_verifications_completed != 0 {
            len += 1;
        }
        if self.veid_verification_score != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgRecordPerformance", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.blocks_proposed != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blocksProposed", ToString::to_string(&self.blocks_proposed).as_str())?;
        }
        if self.blocks_signed != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blocksSigned", ToString::to_string(&self.blocks_signed).as_str())?;
        }
        if self.veid_verifications_completed != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("veidVerificationsCompleted", ToString::to_string(&self.veid_verifications_completed).as_str())?;
        }
        if self.veid_verification_score != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("veidVerificationScore", ToString::to_string(&self.veid_verification_score).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRecordPerformance {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "validator_address",
            "validatorAddress",
            "blocks_proposed",
            "blocksProposed",
            "blocks_signed",
            "blocksSigned",
            "veid_verifications_completed",
            "veidVerificationsCompleted",
            "veid_verification_score",
            "veidVerificationScore",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ValidatorAddress,
            BlocksProposed,
            BlocksSigned,
            VeidVerificationsCompleted,
            VeidVerificationScore,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "blocksProposed" | "blocks_proposed" => Ok(GeneratedField::BlocksProposed),
                            "blocksSigned" | "blocks_signed" => Ok(GeneratedField::BlocksSigned),
                            "veidVerificationsCompleted" | "veid_verifications_completed" => Ok(GeneratedField::VeidVerificationsCompleted),
                            "veidVerificationScore" | "veid_verification_score" => Ok(GeneratedField::VeidVerificationScore),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRecordPerformance;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.MsgRecordPerformance")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRecordPerformance, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut validator_address__ = None;
                let mut blocks_proposed__ = None;
                let mut blocks_signed__ = None;
                let mut veid_verifications_completed__ = None;
                let mut veid_verification_score__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlocksProposed => {
                            if blocks_proposed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blocksProposed"));
                            }
                            blocks_proposed__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlocksSigned => {
                            if blocks_signed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blocksSigned"));
                            }
                            blocks_signed__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidVerificationsCompleted => {
                            if veid_verifications_completed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidVerificationsCompleted"));
                            }
                            veid_verifications_completed__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidVerificationScore => {
                            if veid_verification_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidVerificationScore"));
                            }
                            veid_verification_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRecordPerformance {
                    authority: authority__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    blocks_proposed: blocks_proposed__.unwrap_or_default(),
                    blocks_signed: blocks_signed__.unwrap_or_default(),
                    veid_verifications_completed: veid_verifications_completed__.unwrap_or_default(),
                    veid_verification_score: veid_verification_score__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.MsgRecordPerformance", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRecordPerformanceResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgRecordPerformanceResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRecordPerformanceResponse {
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
            type Value = MsgRecordPerformanceResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.MsgRecordPerformanceResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRecordPerformanceResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRecordPerformanceResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.MsgRecordPerformanceResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSlashValidator {
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
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.reason != 0 {
            len += 1;
        }
        if self.infraction_height != 0 {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgSlashValidator", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.reason != 0 {
            let v = SlashReason::try_from(self.reason)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.reason)))?;
            struct_ser.serialize_field("reason", &v)?;
        }
        if self.infraction_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("infractionHeight", ToString::to_string(&self.infraction_height).as_str())?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSlashValidator {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "validator_address",
            "validatorAddress",
            "reason",
            "infraction_height",
            "infractionHeight",
            "evidence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ValidatorAddress,
            Reason,
            InfractionHeight,
            Evidence,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "reason" => Ok(GeneratedField::Reason),
                            "infractionHeight" | "infraction_height" => Ok(GeneratedField::InfractionHeight),
                            "evidence" => Ok(GeneratedField::Evidence),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSlashValidator;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.MsgSlashValidator")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSlashValidator, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut validator_address__ = None;
                let mut reason__ = None;
                let mut infraction_height__ = None;
                let mut evidence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value::<SlashReason>()? as i32);
                        }
                        GeneratedField::InfractionHeight => {
                            if infraction_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("infractionHeight"));
                            }
                            infraction_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSlashValidator {
                    authority: authority__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    infraction_height: infraction_height__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.MsgSlashValidator", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSlashValidatorResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgSlashValidatorResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSlashValidatorResponse {
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
            type Value = MsgSlashValidatorResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.MsgSlashValidatorResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSlashValidatorResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgSlashValidatorResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.MsgSlashValidatorResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUnjailValidator {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgUnjailValidator", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUnjailValidator {
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
            type Value = MsgUnjailValidator;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.MsgUnjailValidator")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUnjailValidator, V::Error>
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
                Ok(MsgUnjailValidator {
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.MsgUnjailValidator", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUnjailValidatorResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgUnjailValidatorResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUnjailValidatorResponse {
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
            type Value = MsgUnjailValidatorResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.MsgUnjailValidatorResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUnjailValidatorResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUnjailValidatorResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.MsgUnjailValidatorResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.staking.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.staking.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.staking.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.staking.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.staking.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if self.epoch_length != 0 {
            len += 1;
        }
        if self.base_reward_per_block != 0 {
            len += 1;
        }
        if self.veid_reward_pool != 0 {
            len += 1;
        }
        if self.identity_network_reward_pool != 0 {
            len += 1;
        }
        if self.downtime_threshold != 0 {
            len += 1;
        }
        if self.signed_blocks_window != 0 {
            len += 1;
        }
        if self.min_signed_per_window != 0 {
            len += 1;
        }
        if self.slash_fraction_double_sign != 0 {
            len += 1;
        }
        if self.slash_fraction_downtime != 0 {
            len += 1;
        }
        if self.slash_fraction_invalid_attestation != 0 {
            len += 1;
        }
        if self.jail_duration_downtime != 0 {
            len += 1;
        }
        if self.jail_duration_double_sign != 0 {
            len += 1;
        }
        if self.jail_duration_invalid_attestation != 0 {
            len += 1;
        }
        if self.score_tolerance != 0 {
            len += 1;
        }
        if self.max_missed_veid_recomputations != 0 {
            len += 1;
        }
        if !self.reward_denom.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.Params", len)?;
        if self.epoch_length != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochLength", ToString::to_string(&self.epoch_length).as_str())?;
        }
        if self.base_reward_per_block != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("baseRewardPerBlock", ToString::to_string(&self.base_reward_per_block).as_str())?;
        }
        if self.veid_reward_pool != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("veidRewardPool", ToString::to_string(&self.veid_reward_pool).as_str())?;
        }
        if self.identity_network_reward_pool != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("identityNetworkRewardPool", ToString::to_string(&self.identity_network_reward_pool).as_str())?;
        }
        if self.downtime_threshold != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("downtimeThreshold", ToString::to_string(&self.downtime_threshold).as_str())?;
        }
        if self.signed_blocks_window != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("signedBlocksWindow", ToString::to_string(&self.signed_blocks_window).as_str())?;
        }
        if self.min_signed_per_window != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("minSignedPerWindow", ToString::to_string(&self.min_signed_per_window).as_str())?;
        }
        if self.slash_fraction_double_sign != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashFractionDoubleSign", ToString::to_string(&self.slash_fraction_double_sign).as_str())?;
        }
        if self.slash_fraction_downtime != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashFractionDowntime", ToString::to_string(&self.slash_fraction_downtime).as_str())?;
        }
        if self.slash_fraction_invalid_attestation != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashFractionInvalidAttestation", ToString::to_string(&self.slash_fraction_invalid_attestation).as_str())?;
        }
        if self.jail_duration_downtime != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("jailDurationDowntime", ToString::to_string(&self.jail_duration_downtime).as_str())?;
        }
        if self.jail_duration_double_sign != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("jailDurationDoubleSign", ToString::to_string(&self.jail_duration_double_sign).as_str())?;
        }
        if self.jail_duration_invalid_attestation != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("jailDurationInvalidAttestation", ToString::to_string(&self.jail_duration_invalid_attestation).as_str())?;
        }
        if self.score_tolerance != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("scoreTolerance", ToString::to_string(&self.score_tolerance).as_str())?;
        }
        if self.max_missed_veid_recomputations != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxMissedVeidRecomputations", ToString::to_string(&self.max_missed_veid_recomputations).as_str())?;
        }
        if !self.reward_denom.is_empty() {
            struct_ser.serialize_field("rewardDenom", &self.reward_denom)?;
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
            "epoch_length",
            "epochLength",
            "base_reward_per_block",
            "baseRewardPerBlock",
            "veid_reward_pool",
            "veidRewardPool",
            "identity_network_reward_pool",
            "identityNetworkRewardPool",
            "downtime_threshold",
            "downtimeThreshold",
            "signed_blocks_window",
            "signedBlocksWindow",
            "min_signed_per_window",
            "minSignedPerWindow",
            "slash_fraction_double_sign",
            "slashFractionDoubleSign",
            "slash_fraction_downtime",
            "slashFractionDowntime",
            "slash_fraction_invalid_attestation",
            "slashFractionInvalidAttestation",
            "jail_duration_downtime",
            "jailDurationDowntime",
            "jail_duration_double_sign",
            "jailDurationDoubleSign",
            "jail_duration_invalid_attestation",
            "jailDurationInvalidAttestation",
            "score_tolerance",
            "scoreTolerance",
            "max_missed_veid_recomputations",
            "maxMissedVeidRecomputations",
            "reward_denom",
            "rewardDenom",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochLength,
            BaseRewardPerBlock,
            VeidRewardPool,
            IdentityNetworkRewardPool,
            DowntimeThreshold,
            SignedBlocksWindow,
            MinSignedPerWindow,
            SlashFractionDoubleSign,
            SlashFractionDowntime,
            SlashFractionInvalidAttestation,
            JailDurationDowntime,
            JailDurationDoubleSign,
            JailDurationInvalidAttestation,
            ScoreTolerance,
            MaxMissedVeidRecomputations,
            RewardDenom,
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
                            "epochLength" | "epoch_length" => Ok(GeneratedField::EpochLength),
                            "baseRewardPerBlock" | "base_reward_per_block" => Ok(GeneratedField::BaseRewardPerBlock),
                            "veidRewardPool" | "veid_reward_pool" => Ok(GeneratedField::VeidRewardPool),
                            "identityNetworkRewardPool" | "identity_network_reward_pool" => Ok(GeneratedField::IdentityNetworkRewardPool),
                            "downtimeThreshold" | "downtime_threshold" => Ok(GeneratedField::DowntimeThreshold),
                            "signedBlocksWindow" | "signed_blocks_window" => Ok(GeneratedField::SignedBlocksWindow),
                            "minSignedPerWindow" | "min_signed_per_window" => Ok(GeneratedField::MinSignedPerWindow),
                            "slashFractionDoubleSign" | "slash_fraction_double_sign" => Ok(GeneratedField::SlashFractionDoubleSign),
                            "slashFractionDowntime" | "slash_fraction_downtime" => Ok(GeneratedField::SlashFractionDowntime),
                            "slashFractionInvalidAttestation" | "slash_fraction_invalid_attestation" => Ok(GeneratedField::SlashFractionInvalidAttestation),
                            "jailDurationDowntime" | "jail_duration_downtime" => Ok(GeneratedField::JailDurationDowntime),
                            "jailDurationDoubleSign" | "jail_duration_double_sign" => Ok(GeneratedField::JailDurationDoubleSign),
                            "jailDurationInvalidAttestation" | "jail_duration_invalid_attestation" => Ok(GeneratedField::JailDurationInvalidAttestation),
                            "scoreTolerance" | "score_tolerance" => Ok(GeneratedField::ScoreTolerance),
                            "maxMissedVeidRecomputations" | "max_missed_veid_recomputations" => Ok(GeneratedField::MaxMissedVeidRecomputations),
                            "rewardDenom" | "reward_denom" => Ok(GeneratedField::RewardDenom),
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
                formatter.write_str("struct virtengine.staking.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_length__ = None;
                let mut base_reward_per_block__ = None;
                let mut veid_reward_pool__ = None;
                let mut identity_network_reward_pool__ = None;
                let mut downtime_threshold__ = None;
                let mut signed_blocks_window__ = None;
                let mut min_signed_per_window__ = None;
                let mut slash_fraction_double_sign__ = None;
                let mut slash_fraction_downtime__ = None;
                let mut slash_fraction_invalid_attestation__ = None;
                let mut jail_duration_downtime__ = None;
                let mut jail_duration_double_sign__ = None;
                let mut jail_duration_invalid_attestation__ = None;
                let mut score_tolerance__ = None;
                let mut max_missed_veid_recomputations__ = None;
                let mut reward_denom__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EpochLength => {
                            if epoch_length__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochLength"));
                            }
                            epoch_length__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BaseRewardPerBlock => {
                            if base_reward_per_block__.is_some() {
                                return Err(serde::de::Error::duplicate_field("baseRewardPerBlock"));
                            }
                            base_reward_per_block__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidRewardPool => {
                            if veid_reward_pool__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidRewardPool"));
                            }
                            veid_reward_pool__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IdentityNetworkRewardPool => {
                            if identity_network_reward_pool__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identityNetworkRewardPool"));
                            }
                            identity_network_reward_pool__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DowntimeThreshold => {
                            if downtime_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("downtimeThreshold"));
                            }
                            downtime_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SignedBlocksWindow => {
                            if signed_blocks_window__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signedBlocksWindow"));
                            }
                            signed_blocks_window__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinSignedPerWindow => {
                            if min_signed_per_window__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minSignedPerWindow"));
                            }
                            min_signed_per_window__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SlashFractionDoubleSign => {
                            if slash_fraction_double_sign__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashFractionDoubleSign"));
                            }
                            slash_fraction_double_sign__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SlashFractionDowntime => {
                            if slash_fraction_downtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashFractionDowntime"));
                            }
                            slash_fraction_downtime__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SlashFractionInvalidAttestation => {
                            if slash_fraction_invalid_attestation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashFractionInvalidAttestation"));
                            }
                            slash_fraction_invalid_attestation__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JailDurationDowntime => {
                            if jail_duration_downtime__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailDurationDowntime"));
                            }
                            jail_duration_downtime__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JailDurationDoubleSign => {
                            if jail_duration_double_sign__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailDurationDoubleSign"));
                            }
                            jail_duration_double_sign__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JailDurationInvalidAttestation => {
                            if jail_duration_invalid_attestation__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailDurationInvalidAttestation"));
                            }
                            jail_duration_invalid_attestation__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScoreTolerance => {
                            if score_tolerance__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreTolerance"));
                            }
                            score_tolerance__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxMissedVeidRecomputations => {
                            if max_missed_veid_recomputations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxMissedVeidRecomputations"));
                            }
                            max_missed_veid_recomputations__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RewardDenom => {
                            if reward_denom__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rewardDenom"));
                            }
                            reward_denom__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Params {
                    epoch_length: epoch_length__.unwrap_or_default(),
                    base_reward_per_block: base_reward_per_block__.unwrap_or_default(),
                    veid_reward_pool: veid_reward_pool__.unwrap_or_default(),
                    identity_network_reward_pool: identity_network_reward_pool__.unwrap_or_default(),
                    downtime_threshold: downtime_threshold__.unwrap_or_default(),
                    signed_blocks_window: signed_blocks_window__.unwrap_or_default(),
                    min_signed_per_window: min_signed_per_window__.unwrap_or_default(),
                    slash_fraction_double_sign: slash_fraction_double_sign__.unwrap_or_default(),
                    slash_fraction_downtime: slash_fraction_downtime__.unwrap_or_default(),
                    slash_fraction_invalid_attestation: slash_fraction_invalid_attestation__.unwrap_or_default(),
                    jail_duration_downtime: jail_duration_downtime__.unwrap_or_default(),
                    jail_duration_double_sign: jail_duration_double_sign__.unwrap_or_default(),
                    jail_duration_invalid_attestation: jail_duration_invalid_attestation__.unwrap_or_default(),
                    score_tolerance: score_tolerance__.unwrap_or_default(),
                    max_missed_veid_recomputations: max_missed_veid_recomputations__.unwrap_or_default(),
                    reward_denom: reward_denom__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RewardEpoch {
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
        if self.start_height != 0 {
            len += 1;
        }
        if self.end_height != 0 {
            len += 1;
        }
        if self.start_time.is_some() {
            len += 1;
        }
        if self.end_time.is_some() {
            len += 1;
        }
        if !self.total_rewards_distributed.is_empty() {
            len += 1;
        }
        if !self.block_proposal_rewards.is_empty() {
            len += 1;
        }
        if !self.veid_rewards.is_empty() {
            len += 1;
        }
        if !self.uptime_rewards.is_empty() {
            len += 1;
        }
        if self.validator_count != 0 {
            len += 1;
        }
        if !self.total_stake.is_empty() {
            len += 1;
        }
        if self.finalized {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.RewardEpoch", len)?;
        if self.epoch_number != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if self.start_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("startHeight", ToString::to_string(&self.start_height).as_str())?;
        }
        if self.end_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("endHeight", ToString::to_string(&self.end_height).as_str())?;
        }
        if let Some(v) = self.start_time.as_ref() {
            struct_ser.serialize_field("startTime", v)?;
        }
        if let Some(v) = self.end_time.as_ref() {
            struct_ser.serialize_field("endTime", v)?;
        }
        if !self.total_rewards_distributed.is_empty() {
            struct_ser.serialize_field("totalRewardsDistributed", &self.total_rewards_distributed)?;
        }
        if !self.block_proposal_rewards.is_empty() {
            struct_ser.serialize_field("blockProposalRewards", &self.block_proposal_rewards)?;
        }
        if !self.veid_rewards.is_empty() {
            struct_ser.serialize_field("veidRewards", &self.veid_rewards)?;
        }
        if !self.uptime_rewards.is_empty() {
            struct_ser.serialize_field("uptimeRewards", &self.uptime_rewards)?;
        }
        if self.validator_count != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("validatorCount", ToString::to_string(&self.validator_count).as_str())?;
        }
        if !self.total_stake.is_empty() {
            struct_ser.serialize_field("totalStake", &self.total_stake)?;
        }
        if self.finalized {
            struct_ser.serialize_field("finalized", &self.finalized)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RewardEpoch {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_number",
            "epochNumber",
            "start_height",
            "startHeight",
            "end_height",
            "endHeight",
            "start_time",
            "startTime",
            "end_time",
            "endTime",
            "total_rewards_distributed",
            "totalRewardsDistributed",
            "block_proposal_rewards",
            "blockProposalRewards",
            "veid_rewards",
            "veidRewards",
            "uptime_rewards",
            "uptimeRewards",
            "validator_count",
            "validatorCount",
            "total_stake",
            "totalStake",
            "finalized",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNumber,
            StartHeight,
            EndHeight,
            StartTime,
            EndTime,
            TotalRewardsDistributed,
            BlockProposalRewards,
            VeidRewards,
            UptimeRewards,
            ValidatorCount,
            TotalStake,
            Finalized,
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
                            "startHeight" | "start_height" => Ok(GeneratedField::StartHeight),
                            "endHeight" | "end_height" => Ok(GeneratedField::EndHeight),
                            "startTime" | "start_time" => Ok(GeneratedField::StartTime),
                            "endTime" | "end_time" => Ok(GeneratedField::EndTime),
                            "totalRewardsDistributed" | "total_rewards_distributed" => Ok(GeneratedField::TotalRewardsDistributed),
                            "blockProposalRewards" | "block_proposal_rewards" => Ok(GeneratedField::BlockProposalRewards),
                            "veidRewards" | "veid_rewards" => Ok(GeneratedField::VeidRewards),
                            "uptimeRewards" | "uptime_rewards" => Ok(GeneratedField::UptimeRewards),
                            "validatorCount" | "validator_count" => Ok(GeneratedField::ValidatorCount),
                            "totalStake" | "total_stake" => Ok(GeneratedField::TotalStake),
                            "finalized" => Ok(GeneratedField::Finalized),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RewardEpoch;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.RewardEpoch")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<RewardEpoch, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_number__ = None;
                let mut start_height__ = None;
                let mut end_height__ = None;
                let mut start_time__ = None;
                let mut end_time__ = None;
                let mut total_rewards_distributed__ = None;
                let mut block_proposal_rewards__ = None;
                let mut veid_rewards__ = None;
                let mut uptime_rewards__ = None;
                let mut validator_count__ = None;
                let mut total_stake__ = None;
                let mut finalized__ = None;
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
                        GeneratedField::StartHeight => {
                            if start_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("startHeight"));
                            }
                            start_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EndHeight => {
                            if end_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("endHeight"));
                            }
                            end_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::StartTime => {
                            if start_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("startTime"));
                            }
                            start_time__ = map_.next_value()?;
                        }
                        GeneratedField::EndTime => {
                            if end_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("endTime"));
                            }
                            end_time__ = map_.next_value()?;
                        }
                        GeneratedField::TotalRewardsDistributed => {
                            if total_rewards_distributed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalRewardsDistributed"));
                            }
                            total_rewards_distributed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlockProposalRewards => {
                            if block_proposal_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockProposalRewards"));
                            }
                            block_proposal_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::VeidRewards => {
                            if veid_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidRewards"));
                            }
                            veid_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UptimeRewards => {
                            if uptime_rewards__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uptimeRewards"));
                            }
                            uptime_rewards__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorCount => {
                            if validator_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorCount"));
                            }
                            validator_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TotalStake => {
                            if total_stake__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalStake"));
                            }
                            total_stake__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Finalized => {
                            if finalized__.is_some() {
                                return Err(serde::de::Error::duplicate_field("finalized"));
                            }
                            finalized__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(RewardEpoch {
                    epoch_number: epoch_number__.unwrap_or_default(),
                    start_height: start_height__.unwrap_or_default(),
                    end_height: end_height__.unwrap_or_default(),
                    start_time: start_time__,
                    end_time: end_time__,
                    total_rewards_distributed: total_rewards_distributed__.unwrap_or_default(),
                    block_proposal_rewards: block_proposal_rewards__.unwrap_or_default(),
                    veid_rewards: veid_rewards__.unwrap_or_default(),
                    uptime_rewards: uptime_rewards__.unwrap_or_default(),
                    validator_count: validator_count__.unwrap_or_default(),
                    total_stake: total_stake__.unwrap_or_default(),
                    finalized: finalized__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.RewardEpoch", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RewardType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "REWARD_TYPE_UNSPECIFIED",
            Self::BlockProposal => "REWARD_TYPE_BLOCK_PROPOSAL",
            Self::VeidVerification => "REWARD_TYPE_VEID_VERIFICATION",
            Self::Uptime => "REWARD_TYPE_UPTIME",
            Self::IdentityNetwork => "REWARD_TYPE_IDENTITY_NETWORK",
            Self::Staking => "REWARD_TYPE_STAKING",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for RewardType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "REWARD_TYPE_UNSPECIFIED",
            "REWARD_TYPE_BLOCK_PROPOSAL",
            "REWARD_TYPE_VEID_VERIFICATION",
            "REWARD_TYPE_UPTIME",
            "REWARD_TYPE_IDENTITY_NETWORK",
            "REWARD_TYPE_STAKING",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RewardType;

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
                    "REWARD_TYPE_UNSPECIFIED" => Ok(RewardType::Unspecified),
                    "REWARD_TYPE_BLOCK_PROPOSAL" => Ok(RewardType::BlockProposal),
                    "REWARD_TYPE_VEID_VERIFICATION" => Ok(RewardType::VeidVerification),
                    "REWARD_TYPE_UPTIME" => Ok(RewardType::Uptime),
                    "REWARD_TYPE_IDENTITY_NETWORK" => Ok(RewardType::IdentityNetwork),
                    "REWARD_TYPE_STAKING" => Ok(RewardType::Staking),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for SlashConfig {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.reason != 0 {
            len += 1;
        }
        if self.slash_percent != 0 {
            len += 1;
        }
        if self.jail_duration != 0 {
            len += 1;
        }
        if self.is_tombstone {
            len += 1;
        }
        if self.escalation_multiplier != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.SlashConfig", len)?;
        if self.reason != 0 {
            let v = SlashReason::try_from(self.reason)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.reason)))?;
            struct_ser.serialize_field("reason", &v)?;
        }
        if self.slash_percent != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashPercent", ToString::to_string(&self.slash_percent).as_str())?;
        }
        if self.jail_duration != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("jailDuration", ToString::to_string(&self.jail_duration).as_str())?;
        }
        if self.is_tombstone {
            struct_ser.serialize_field("isTombstone", &self.is_tombstone)?;
        }
        if self.escalation_multiplier != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("escalationMultiplier", ToString::to_string(&self.escalation_multiplier).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SlashConfig {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reason",
            "slash_percent",
            "slashPercent",
            "jail_duration",
            "jailDuration",
            "is_tombstone",
            "isTombstone",
            "escalation_multiplier",
            "escalationMultiplier",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reason,
            SlashPercent,
            JailDuration,
            IsTombstone,
            EscalationMultiplier,
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
                            "reason" => Ok(GeneratedField::Reason),
                            "slashPercent" | "slash_percent" => Ok(GeneratedField::SlashPercent),
                            "jailDuration" | "jail_duration" => Ok(GeneratedField::JailDuration),
                            "isTombstone" | "is_tombstone" => Ok(GeneratedField::IsTombstone),
                            "escalationMultiplier" | "escalation_multiplier" => Ok(GeneratedField::EscalationMultiplier),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SlashConfig;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.SlashConfig")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SlashConfig, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reason__ = None;
                let mut slash_percent__ = None;
                let mut jail_duration__ = None;
                let mut is_tombstone__ = None;
                let mut escalation_multiplier__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value::<SlashReason>()? as i32);
                        }
                        GeneratedField::SlashPercent => {
                            if slash_percent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashPercent"));
                            }
                            slash_percent__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JailDuration => {
                            if jail_duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailDuration"));
                            }
                            jail_duration__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsTombstone => {
                            if is_tombstone__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isTombstone"));
                            }
                            is_tombstone__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscalationMultiplier => {
                            if escalation_multiplier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escalationMultiplier"));
                            }
                            escalation_multiplier__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(SlashConfig {
                    reason: reason__.unwrap_or_default(),
                    slash_percent: slash_percent__.unwrap_or_default(),
                    jail_duration: jail_duration__.unwrap_or_default(),
                    is_tombstone: is_tombstone__.unwrap_or_default(),
                    escalation_multiplier: escalation_multiplier__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.SlashConfig", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SlashReason {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "SLASH_REASON_UNSPECIFIED",
            Self::DoubleSigning => "SLASH_REASON_DOUBLE_SIGNING",
            Self::Downtime => "SLASH_REASON_DOWNTIME",
            Self::InvalidVeidAttestation => "SLASH_REASON_INVALID_VEID_ATTESTATION",
            Self::MissedRecomputation => "SLASH_REASON_MISSED_RECOMPUTATION",
            Self::InconsistentScore => "SLASH_REASON_INCONSISTENT_SCORE",
            Self::ExpiredAttestation => "SLASH_REASON_EXPIRED_ATTESTATION",
            Self::DebugModeEnabled => "SLASH_REASON_DEBUG_MODE_ENABLED",
            Self::NonAllowlistedMeasurement => "SLASH_REASON_NON_ALLOWLISTED_MEASUREMENT",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for SlashReason {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "SLASH_REASON_UNSPECIFIED",
            "SLASH_REASON_DOUBLE_SIGNING",
            "SLASH_REASON_DOWNTIME",
            "SLASH_REASON_INVALID_VEID_ATTESTATION",
            "SLASH_REASON_MISSED_RECOMPUTATION",
            "SLASH_REASON_INCONSISTENT_SCORE",
            "SLASH_REASON_EXPIRED_ATTESTATION",
            "SLASH_REASON_DEBUG_MODE_ENABLED",
            "SLASH_REASON_NON_ALLOWLISTED_MEASUREMENT",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SlashReason;

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
                    "SLASH_REASON_UNSPECIFIED" => Ok(SlashReason::Unspecified),
                    "SLASH_REASON_DOUBLE_SIGNING" => Ok(SlashReason::DoubleSigning),
                    "SLASH_REASON_DOWNTIME" => Ok(SlashReason::Downtime),
                    "SLASH_REASON_INVALID_VEID_ATTESTATION" => Ok(SlashReason::InvalidVeidAttestation),
                    "SLASH_REASON_MISSED_RECOMPUTATION" => Ok(SlashReason::MissedRecomputation),
                    "SLASH_REASON_INCONSISTENT_SCORE" => Ok(SlashReason::InconsistentScore),
                    "SLASH_REASON_EXPIRED_ATTESTATION" => Ok(SlashReason::ExpiredAttestation),
                    "SLASH_REASON_DEBUG_MODE_ENABLED" => Ok(SlashReason::DebugModeEnabled),
                    "SLASH_REASON_NON_ALLOWLISTED_MEASUREMENT" => Ok(SlashReason::NonAllowlistedMeasurement),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for SlashRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.slash_id.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if self.reason != 0 {
            len += 1;
        }
        if !self.amount.is_empty() {
            len += 1;
        }
        if self.slash_percent != 0 {
            len += 1;
        }
        if self.infraction_height != 0 {
            len += 1;
        }
        if self.slash_height != 0 {
            len += 1;
        }
        if self.slash_time.is_some() {
            len += 1;
        }
        if self.jailed {
            len += 1;
        }
        if self.jail_duration != 0 {
            len += 1;
        }
        if self.jailed_until.is_some() {
            len += 1;
        }
        if self.tombstoned {
            len += 1;
        }
        if !self.evidence.is_empty() {
            len += 1;
        }
        if !self.evidence_hash.is_empty() {
            len += 1;
        }
        if !self.reporter_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.SlashRecord", len)?;
        if !self.slash_id.is_empty() {
            struct_ser.serialize_field("slashId", &self.slash_id)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.reason != 0 {
            let v = SlashReason::try_from(self.reason)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.reason)))?;
            struct_ser.serialize_field("reason", &v)?;
        }
        if !self.amount.is_empty() {
            struct_ser.serialize_field("amount", &self.amount)?;
        }
        if self.slash_percent != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashPercent", ToString::to_string(&self.slash_percent).as_str())?;
        }
        if self.infraction_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("infractionHeight", ToString::to_string(&self.infraction_height).as_str())?;
        }
        if self.slash_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("slashHeight", ToString::to_string(&self.slash_height).as_str())?;
        }
        if let Some(v) = self.slash_time.as_ref() {
            struct_ser.serialize_field("slashTime", v)?;
        }
        if self.jailed {
            struct_ser.serialize_field("jailed", &self.jailed)?;
        }
        if self.jail_duration != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("jailDuration", ToString::to_string(&self.jail_duration).as_str())?;
        }
        if let Some(v) = self.jailed_until.as_ref() {
            struct_ser.serialize_field("jailedUntil", v)?;
        }
        if self.tombstoned {
            struct_ser.serialize_field("tombstoned", &self.tombstoned)?;
        }
        if !self.evidence.is_empty() {
            struct_ser.serialize_field("evidence", &self.evidence)?;
        }
        if !self.evidence_hash.is_empty() {
            struct_ser.serialize_field("evidenceHash", &self.evidence_hash)?;
        }
        if !self.reporter_address.is_empty() {
            struct_ser.serialize_field("reporterAddress", &self.reporter_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SlashRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "slash_id",
            "slashId",
            "validator_address",
            "validatorAddress",
            "reason",
            "amount",
            "slash_percent",
            "slashPercent",
            "infraction_height",
            "infractionHeight",
            "slash_height",
            "slashHeight",
            "slash_time",
            "slashTime",
            "jailed",
            "jail_duration",
            "jailDuration",
            "jailed_until",
            "jailedUntil",
            "tombstoned",
            "evidence",
            "evidence_hash",
            "evidenceHash",
            "reporter_address",
            "reporterAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            SlashId,
            ValidatorAddress,
            Reason,
            Amount,
            SlashPercent,
            InfractionHeight,
            SlashHeight,
            SlashTime,
            Jailed,
            JailDuration,
            JailedUntil,
            Tombstoned,
            Evidence,
            EvidenceHash,
            ReporterAddress,
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
                            "slashId" | "slash_id" => Ok(GeneratedField::SlashId),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "reason" => Ok(GeneratedField::Reason),
                            "amount" => Ok(GeneratedField::Amount),
                            "slashPercent" | "slash_percent" => Ok(GeneratedField::SlashPercent),
                            "infractionHeight" | "infraction_height" => Ok(GeneratedField::InfractionHeight),
                            "slashHeight" | "slash_height" => Ok(GeneratedField::SlashHeight),
                            "slashTime" | "slash_time" => Ok(GeneratedField::SlashTime),
                            "jailed" => Ok(GeneratedField::Jailed),
                            "jailDuration" | "jail_duration" => Ok(GeneratedField::JailDuration),
                            "jailedUntil" | "jailed_until" => Ok(GeneratedField::JailedUntil),
                            "tombstoned" => Ok(GeneratedField::Tombstoned),
                            "evidence" => Ok(GeneratedField::Evidence),
                            "evidenceHash" | "evidence_hash" => Ok(GeneratedField::EvidenceHash),
                            "reporterAddress" | "reporter_address" => Ok(GeneratedField::ReporterAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SlashRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.SlashRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<SlashRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut slash_id__ = None;
                let mut validator_address__ = None;
                let mut reason__ = None;
                let mut amount__ = None;
                let mut slash_percent__ = None;
                let mut infraction_height__ = None;
                let mut slash_height__ = None;
                let mut slash_time__ = None;
                let mut jailed__ = None;
                let mut jail_duration__ = None;
                let mut jailed_until__ = None;
                let mut tombstoned__ = None;
                let mut evidence__ = None;
                let mut evidence_hash__ = None;
                let mut reporter_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::SlashId => {
                            if slash_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashId"));
                            }
                            slash_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value::<SlashReason>()? as i32);
                        }
                        GeneratedField::Amount => {
                            if amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("amount"));
                            }
                            amount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SlashPercent => {
                            if slash_percent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashPercent"));
                            }
                            slash_percent__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::InfractionHeight => {
                            if infraction_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("infractionHeight"));
                            }
                            infraction_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SlashHeight => {
                            if slash_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashHeight"));
                            }
                            slash_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SlashTime => {
                            if slash_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("slashTime"));
                            }
                            slash_time__ = map_.next_value()?;
                        }
                        GeneratedField::Jailed => {
                            if jailed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailed"));
                            }
                            jailed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::JailDuration => {
                            if jail_duration__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailDuration"));
                            }
                            jail_duration__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JailedUntil => {
                            if jailed_until__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailedUntil"));
                            }
                            jailed_until__ = map_.next_value()?;
                        }
                        GeneratedField::Tombstoned => {
                            if tombstoned__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tombstoned"));
                            }
                            tombstoned__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Evidence => {
                            if evidence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidence"));
                            }
                            evidence__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EvidenceHash => {
                            if evidence_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidenceHash"));
                            }
                            evidence_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReporterAddress => {
                            if reporter_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporterAddress"));
                            }
                            reporter_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(SlashRecord {
                    slash_id: slash_id__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    amount: amount__.unwrap_or_default(),
                    slash_percent: slash_percent__.unwrap_or_default(),
                    infraction_height: infraction_height__.unwrap_or_default(),
                    slash_height: slash_height__.unwrap_or_default(),
                    slash_time: slash_time__,
                    jailed: jailed__.unwrap_or_default(),
                    jail_duration: jail_duration__.unwrap_or_default(),
                    jailed_until: jailed_until__,
                    tombstoned: tombstoned__.unwrap_or_default(),
                    evidence: evidence__.unwrap_or_default(),
                    evidence_hash: evidence_hash__.unwrap_or_default(),
                    reporter_address: reporter_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.SlashRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorPerformance {
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
        if self.blocks_proposed != 0 {
            len += 1;
        }
        if self.blocks_expected != 0 {
            len += 1;
        }
        if self.blocks_missed != 0 {
            len += 1;
        }
        if self.total_signatures != 0 {
            len += 1;
        }
        if self.veid_verifications_completed != 0 {
            len += 1;
        }
        if self.veid_verifications_expected != 0 {
            len += 1;
        }
        if self.veid_verification_score != 0 {
            len += 1;
        }
        if self.uptime_seconds != 0 {
            len += 1;
        }
        if self.downtime_seconds != 0 {
            len += 1;
        }
        if self.consecutive_missed_blocks != 0 {
            len += 1;
        }
        if self.last_proposed_height != 0 {
            len += 1;
        }
        if self.last_signed_height != 0 {
            len += 1;
        }
        if self.epoch_number != 0 {
            len += 1;
        }
        if self.updated_at.is_some() {
            len += 1;
        }
        if self.overall_score != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.ValidatorPerformance", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.blocks_proposed != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blocksProposed", ToString::to_string(&self.blocks_proposed).as_str())?;
        }
        if self.blocks_expected != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blocksExpected", ToString::to_string(&self.blocks_expected).as_str())?;
        }
        if self.blocks_missed != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blocksMissed", ToString::to_string(&self.blocks_missed).as_str())?;
        }
        if self.total_signatures != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("totalSignatures", ToString::to_string(&self.total_signatures).as_str())?;
        }
        if self.veid_verifications_completed != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("veidVerificationsCompleted", ToString::to_string(&self.veid_verifications_completed).as_str())?;
        }
        if self.veid_verifications_expected != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("veidVerificationsExpected", ToString::to_string(&self.veid_verifications_expected).as_str())?;
        }
        if self.veid_verification_score != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("veidVerificationScore", ToString::to_string(&self.veid_verification_score).as_str())?;
        }
        if self.uptime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("uptimeSeconds", ToString::to_string(&self.uptime_seconds).as_str())?;
        }
        if self.downtime_seconds != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("downtimeSeconds", ToString::to_string(&self.downtime_seconds).as_str())?;
        }
        if self.consecutive_missed_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("consecutiveMissedBlocks", ToString::to_string(&self.consecutive_missed_blocks).as_str())?;
        }
        if self.last_proposed_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastProposedHeight", ToString::to_string(&self.last_proposed_height).as_str())?;
        }
        if self.last_signed_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastSignedHeight", ToString::to_string(&self.last_signed_height).as_str())?;
        }
        if self.epoch_number != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if let Some(v) = self.updated_at.as_ref() {
            struct_ser.serialize_field("updatedAt", v)?;
        }
        if self.overall_score != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("overallScore", ToString::to_string(&self.overall_score).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValidatorPerformance {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "blocks_proposed",
            "blocksProposed",
            "blocks_expected",
            "blocksExpected",
            "blocks_missed",
            "blocksMissed",
            "total_signatures",
            "totalSignatures",
            "veid_verifications_completed",
            "veidVerificationsCompleted",
            "veid_verifications_expected",
            "veidVerificationsExpected",
            "veid_verification_score",
            "veidVerificationScore",
            "uptime_seconds",
            "uptimeSeconds",
            "downtime_seconds",
            "downtimeSeconds",
            "consecutive_missed_blocks",
            "consecutiveMissedBlocks",
            "last_proposed_height",
            "lastProposedHeight",
            "last_signed_height",
            "lastSignedHeight",
            "epoch_number",
            "epochNumber",
            "updated_at",
            "updatedAt",
            "overall_score",
            "overallScore",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            BlocksProposed,
            BlocksExpected,
            BlocksMissed,
            TotalSignatures,
            VeidVerificationsCompleted,
            VeidVerificationsExpected,
            VeidVerificationScore,
            UptimeSeconds,
            DowntimeSeconds,
            ConsecutiveMissedBlocks,
            LastProposedHeight,
            LastSignedHeight,
            EpochNumber,
            UpdatedAt,
            OverallScore,
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
                            "blocksProposed" | "blocks_proposed" => Ok(GeneratedField::BlocksProposed),
                            "blocksExpected" | "blocks_expected" => Ok(GeneratedField::BlocksExpected),
                            "blocksMissed" | "blocks_missed" => Ok(GeneratedField::BlocksMissed),
                            "totalSignatures" | "total_signatures" => Ok(GeneratedField::TotalSignatures),
                            "veidVerificationsCompleted" | "veid_verifications_completed" => Ok(GeneratedField::VeidVerificationsCompleted),
                            "veidVerificationsExpected" | "veid_verifications_expected" => Ok(GeneratedField::VeidVerificationsExpected),
                            "veidVerificationScore" | "veid_verification_score" => Ok(GeneratedField::VeidVerificationScore),
                            "uptimeSeconds" | "uptime_seconds" => Ok(GeneratedField::UptimeSeconds),
                            "downtimeSeconds" | "downtime_seconds" => Ok(GeneratedField::DowntimeSeconds),
                            "consecutiveMissedBlocks" | "consecutive_missed_blocks" => Ok(GeneratedField::ConsecutiveMissedBlocks),
                            "lastProposedHeight" | "last_proposed_height" => Ok(GeneratedField::LastProposedHeight),
                            "lastSignedHeight" | "last_signed_height" => Ok(GeneratedField::LastSignedHeight),
                            "epochNumber" | "epoch_number" => Ok(GeneratedField::EpochNumber),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "overallScore" | "overall_score" => Ok(GeneratedField::OverallScore),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValidatorPerformance;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.ValidatorPerformance")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValidatorPerformance, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut blocks_proposed__ = None;
                let mut blocks_expected__ = None;
                let mut blocks_missed__ = None;
                let mut total_signatures__ = None;
                let mut veid_verifications_completed__ = None;
                let mut veid_verifications_expected__ = None;
                let mut veid_verification_score__ = None;
                let mut uptime_seconds__ = None;
                let mut downtime_seconds__ = None;
                let mut consecutive_missed_blocks__ = None;
                let mut last_proposed_height__ = None;
                let mut last_signed_height__ = None;
                let mut epoch_number__ = None;
                let mut updated_at__ = None;
                let mut overall_score__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlocksProposed => {
                            if blocks_proposed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blocksProposed"));
                            }
                            blocks_proposed__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlocksExpected => {
                            if blocks_expected__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blocksExpected"));
                            }
                            blocks_expected__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlocksMissed => {
                            if blocks_missed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blocksMissed"));
                            }
                            blocks_missed__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::TotalSignatures => {
                            if total_signatures__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalSignatures"));
                            }
                            total_signatures__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidVerificationsCompleted => {
                            if veid_verifications_completed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidVerificationsCompleted"));
                            }
                            veid_verifications_completed__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidVerificationsExpected => {
                            if veid_verifications_expected__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidVerificationsExpected"));
                            }
                            veid_verifications_expected__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VeidVerificationScore => {
                            if veid_verification_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidVerificationScore"));
                            }
                            veid_verification_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UptimeSeconds => {
                            if uptime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uptimeSeconds"));
                            }
                            uptime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DowntimeSeconds => {
                            if downtime_seconds__.is_some() {
                                return Err(serde::de::Error::duplicate_field("downtimeSeconds"));
                            }
                            downtime_seconds__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ConsecutiveMissedBlocks => {
                            if consecutive_missed_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consecutiveMissedBlocks"));
                            }
                            consecutive_missed_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastProposedHeight => {
                            if last_proposed_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastProposedHeight"));
                            }
                            last_proposed_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastSignedHeight => {
                            if last_signed_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastSignedHeight"));
                            }
                            last_signed_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EpochNumber => {
                            if epoch_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNumber"));
                            }
                            epoch_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = map_.next_value()?;
                        }
                        GeneratedField::OverallScore => {
                            if overall_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("overallScore"));
                            }
                            overall_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ValidatorPerformance {
                    validator_address: validator_address__.unwrap_or_default(),
                    blocks_proposed: blocks_proposed__.unwrap_or_default(),
                    blocks_expected: blocks_expected__.unwrap_or_default(),
                    blocks_missed: blocks_missed__.unwrap_or_default(),
                    total_signatures: total_signatures__.unwrap_or_default(),
                    veid_verifications_completed: veid_verifications_completed__.unwrap_or_default(),
                    veid_verifications_expected: veid_verifications_expected__.unwrap_or_default(),
                    veid_verification_score: veid_verification_score__.unwrap_or_default(),
                    uptime_seconds: uptime_seconds__.unwrap_or_default(),
                    downtime_seconds: downtime_seconds__.unwrap_or_default(),
                    consecutive_missed_blocks: consecutive_missed_blocks__.unwrap_or_default(),
                    last_proposed_height: last_proposed_height__.unwrap_or_default(),
                    last_signed_height: last_signed_height__.unwrap_or_default(),
                    epoch_number: epoch_number__.unwrap_or_default(),
                    updated_at: updated_at__,
                    overall_score: overall_score__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.ValidatorPerformance", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorReward {
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
        if self.epoch_number != 0 {
            len += 1;
        }
        if !self.total_reward.is_empty() {
            len += 1;
        }
        if !self.block_proposal_reward.is_empty() {
            len += 1;
        }
        if !self.veid_reward.is_empty() {
            len += 1;
        }
        if !self.uptime_reward.is_empty() {
            len += 1;
        }
        if !self.identity_network_reward.is_empty() {
            len += 1;
        }
        if self.performance_score != 0 {
            len += 1;
        }
        if !self.stake_weight.is_empty() {
            len += 1;
        }
        if self.calculated_at.is_some() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.claimed {
            len += 1;
        }
        if self.claimed_at.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.ValidatorReward", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.epoch_number != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if !self.total_reward.is_empty() {
            struct_ser.serialize_field("totalReward", &self.total_reward)?;
        }
        if !self.block_proposal_reward.is_empty() {
            struct_ser.serialize_field("blockProposalReward", &self.block_proposal_reward)?;
        }
        if !self.veid_reward.is_empty() {
            struct_ser.serialize_field("veidReward", &self.veid_reward)?;
        }
        if !self.uptime_reward.is_empty() {
            struct_ser.serialize_field("uptimeReward", &self.uptime_reward)?;
        }
        if !self.identity_network_reward.is_empty() {
            struct_ser.serialize_field("identityNetworkReward", &self.identity_network_reward)?;
        }
        if self.performance_score != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("performanceScore", ToString::to_string(&self.performance_score).as_str())?;
        }
        if !self.stake_weight.is_empty() {
            struct_ser.serialize_field("stakeWeight", &self.stake_weight)?;
        }
        if let Some(v) = self.calculated_at.as_ref() {
            struct_ser.serialize_field("calculatedAt", v)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
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
impl<'de> serde::Deserialize<'de> for ValidatorReward {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "epoch_number",
            "epochNumber",
            "total_reward",
            "totalReward",
            "block_proposal_reward",
            "blockProposalReward",
            "veid_reward",
            "veidReward",
            "uptime_reward",
            "uptimeReward",
            "identity_network_reward",
            "identityNetworkReward",
            "performance_score",
            "performanceScore",
            "stake_weight",
            "stakeWeight",
            "calculated_at",
            "calculatedAt",
            "block_height",
            "blockHeight",
            "claimed",
            "claimed_at",
            "claimedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            EpochNumber,
            TotalReward,
            BlockProposalReward,
            VeidReward,
            UptimeReward,
            IdentityNetworkReward,
            PerformanceScore,
            StakeWeight,
            CalculatedAt,
            BlockHeight,
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
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
                            "epochNumber" | "epoch_number" => Ok(GeneratedField::EpochNumber),
                            "totalReward" | "total_reward" => Ok(GeneratedField::TotalReward),
                            "blockProposalReward" | "block_proposal_reward" => Ok(GeneratedField::BlockProposalReward),
                            "veidReward" | "veid_reward" => Ok(GeneratedField::VeidReward),
                            "uptimeReward" | "uptime_reward" => Ok(GeneratedField::UptimeReward),
                            "identityNetworkReward" | "identity_network_reward" => Ok(GeneratedField::IdentityNetworkReward),
                            "performanceScore" | "performance_score" => Ok(GeneratedField::PerformanceScore),
                            "stakeWeight" | "stake_weight" => Ok(GeneratedField::StakeWeight),
                            "calculatedAt" | "calculated_at" => Ok(GeneratedField::CalculatedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
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
            type Value = ValidatorReward;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.ValidatorReward")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValidatorReward, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut epoch_number__ = None;
                let mut total_reward__ = None;
                let mut block_proposal_reward__ = None;
                let mut veid_reward__ = None;
                let mut uptime_reward__ = None;
                let mut identity_network_reward__ = None;
                let mut performance_score__ = None;
                let mut stake_weight__ = None;
                let mut calculated_at__ = None;
                let mut block_height__ = None;
                let mut claimed__ = None;
                let mut claimed_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
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
                        GeneratedField::TotalReward => {
                            if total_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalReward"));
                            }
                            total_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlockProposalReward => {
                            if block_proposal_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockProposalReward"));
                            }
                            block_proposal_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::VeidReward => {
                            if veid_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("veidReward"));
                            }
                            veid_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UptimeReward => {
                            if uptime_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uptimeReward"));
                            }
                            uptime_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IdentityNetworkReward => {
                            if identity_network_reward__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identityNetworkReward"));
                            }
                            identity_network_reward__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PerformanceScore => {
                            if performance_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("performanceScore"));
                            }
                            performance_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::StakeWeight => {
                            if stake_weight__.is_some() {
                                return Err(serde::de::Error::duplicate_field("stakeWeight"));
                            }
                            stake_weight__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CalculatedAt => {
                            if calculated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("calculatedAt"));
                            }
                            calculated_at__ = map_.next_value()?;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
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
                Ok(ValidatorReward {
                    validator_address: validator_address__.unwrap_or_default(),
                    epoch_number: epoch_number__.unwrap_or_default(),
                    total_reward: total_reward__.unwrap_or_default(),
                    block_proposal_reward: block_proposal_reward__.unwrap_or_default(),
                    veid_reward: veid_reward__.unwrap_or_default(),
                    uptime_reward: uptime_reward__.unwrap_or_default(),
                    identity_network_reward: identity_network_reward__.unwrap_or_default(),
                    performance_score: performance_score__.unwrap_or_default(),
                    stake_weight: stake_weight__.unwrap_or_default(),
                    calculated_at: calculated_at__,
                    block_height: block_height__.unwrap_or_default(),
                    claimed: claimed__.unwrap_or_default(),
                    claimed_at: claimed_at__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.ValidatorReward", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorSigningInfo {
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
        if self.start_height != 0 {
            len += 1;
        }
        if self.index_offset != 0 {
            len += 1;
        }
        if self.jailed_until.is_some() {
            len += 1;
        }
        if self.tombstoned {
            len += 1;
        }
        if self.missed_blocks_counter != 0 {
            len += 1;
        }
        if self.infraction_count != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.staking.v1.ValidatorSigningInfo", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.start_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("startHeight", ToString::to_string(&self.start_height).as_str())?;
        }
        if self.index_offset != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("indexOffset", ToString::to_string(&self.index_offset).as_str())?;
        }
        if let Some(v) = self.jailed_until.as_ref() {
            struct_ser.serialize_field("jailedUntil", v)?;
        }
        if self.tombstoned {
            struct_ser.serialize_field("tombstoned", &self.tombstoned)?;
        }
        if self.missed_blocks_counter != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("missedBlocksCounter", ToString::to_string(&self.missed_blocks_counter).as_str())?;
        }
        if self.infraction_count != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("infractionCount", ToString::to_string(&self.infraction_count).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValidatorSigningInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "start_height",
            "startHeight",
            "index_offset",
            "indexOffset",
            "jailed_until",
            "jailedUntil",
            "tombstoned",
            "missed_blocks_counter",
            "missedBlocksCounter",
            "infraction_count",
            "infractionCount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            StartHeight,
            IndexOffset,
            JailedUntil,
            Tombstoned,
            MissedBlocksCounter,
            InfractionCount,
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
                            "startHeight" | "start_height" => Ok(GeneratedField::StartHeight),
                            "indexOffset" | "index_offset" => Ok(GeneratedField::IndexOffset),
                            "jailedUntil" | "jailed_until" => Ok(GeneratedField::JailedUntil),
                            "tombstoned" => Ok(GeneratedField::Tombstoned),
                            "missedBlocksCounter" | "missed_blocks_counter" => Ok(GeneratedField::MissedBlocksCounter),
                            "infractionCount" | "infraction_count" => Ok(GeneratedField::InfractionCount),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValidatorSigningInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.staking.v1.ValidatorSigningInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValidatorSigningInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut start_height__ = None;
                let mut index_offset__ = None;
                let mut jailed_until__ = None;
                let mut tombstoned__ = None;
                let mut missed_blocks_counter__ = None;
                let mut infraction_count__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::StartHeight => {
                            if start_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("startHeight"));
                            }
                            start_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IndexOffset => {
                            if index_offset__.is_some() {
                                return Err(serde::de::Error::duplicate_field("indexOffset"));
                            }
                            index_offset__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::JailedUntil => {
                            if jailed_until__.is_some() {
                                return Err(serde::de::Error::duplicate_field("jailedUntil"));
                            }
                            jailed_until__ = map_.next_value()?;
                        }
                        GeneratedField::Tombstoned => {
                            if tombstoned__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tombstoned"));
                            }
                            tombstoned__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MissedBlocksCounter => {
                            if missed_blocks_counter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("missedBlocksCounter"));
                            }
                            missed_blocks_counter__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::InfractionCount => {
                            if infraction_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("infractionCount"));
                            }
                            infraction_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ValidatorSigningInfo {
                    validator_address: validator_address__.unwrap_or_default(),
                    start_height: start_height__.unwrap_or_default(),
                    index_offset: index_offset__.unwrap_or_default(),
                    jailed_until: jailed_until__,
                    tombstoned: tombstoned__.unwrap_or_default(),
                    missed_blocks_counter: missed_blocks_counter__.unwrap_or_default(),
                    infraction_count: infraction_count__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.staking.v1.ValidatorSigningInfo", FIELDS, GeneratedVisitor)
    }
}

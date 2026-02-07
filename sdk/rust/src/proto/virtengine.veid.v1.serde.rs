// @generated
impl serde::Serialize for AccountStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unknown => "ACCOUNT_STATUS_UNKNOWN",
            Self::Pending => "ACCOUNT_STATUS_PENDING",
            Self::InProgress => "ACCOUNT_STATUS_IN_PROGRESS",
            Self::Verified => "ACCOUNT_STATUS_VERIFIED",
            Self::Rejected => "ACCOUNT_STATUS_REJECTED",
            Self::Expired => "ACCOUNT_STATUS_EXPIRED",
            Self::NeedsAdditionalFactor => "ACCOUNT_STATUS_NEEDS_ADDITIONAL_FACTOR",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for AccountStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ACCOUNT_STATUS_UNKNOWN",
            "ACCOUNT_STATUS_PENDING",
            "ACCOUNT_STATUS_IN_PROGRESS",
            "ACCOUNT_STATUS_VERIFIED",
            "ACCOUNT_STATUS_REJECTED",
            "ACCOUNT_STATUS_EXPIRED",
            "ACCOUNT_STATUS_NEEDS_ADDITIONAL_FACTOR",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AccountStatus;

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
                    "ACCOUNT_STATUS_UNKNOWN" => Ok(AccountStatus::Unknown),
                    "ACCOUNT_STATUS_PENDING" => Ok(AccountStatus::Pending),
                    "ACCOUNT_STATUS_IN_PROGRESS" => Ok(AccountStatus::InProgress),
                    "ACCOUNT_STATUS_VERIFIED" => Ok(AccountStatus::Verified),
                    "ACCOUNT_STATUS_REJECTED" => Ok(AccountStatus::Rejected),
                    "ACCOUNT_STATUS_EXPIRED" => Ok(AccountStatus::Expired),
                    "ACCOUNT_STATUS_NEEDS_ADDITIONAL_FACTOR" => Ok(AccountStatus::NeedsAdditionalFactor),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for AppealParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.appeal_window_blocks != 0 {
            len += 1;
        }
        if self.max_appeals_per_scope != 0 {
            len += 1;
        }
        if self.min_appeal_reason_length != 0 {
            len += 1;
        }
        if self.review_timeout_blocks != 0 {
            len += 1;
        }
        if self.enabled {
            len += 1;
        }
        if self.require_escrow_deposit {
            len += 1;
        }
        if self.escrow_deposit_amount != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.AppealParams", len)?;
        if self.appeal_window_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("appealWindowBlocks", ToString::to_string(&self.appeal_window_blocks).as_str())?;
        }
        if self.max_appeals_per_scope != 0 {
            struct_ser.serialize_field("maxAppealsPerScope", &self.max_appeals_per_scope)?;
        }
        if self.min_appeal_reason_length != 0 {
            struct_ser.serialize_field("minAppealReasonLength", &self.min_appeal_reason_length)?;
        }
        if self.review_timeout_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reviewTimeoutBlocks", ToString::to_string(&self.review_timeout_blocks).as_str())?;
        }
        if self.enabled {
            struct_ser.serialize_field("enabled", &self.enabled)?;
        }
        if self.require_escrow_deposit {
            struct_ser.serialize_field("requireEscrowDeposit", &self.require_escrow_deposit)?;
        }
        if self.escrow_deposit_amount != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("escrowDepositAmount", ToString::to_string(&self.escrow_deposit_amount).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AppealParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_window_blocks",
            "appealWindowBlocks",
            "max_appeals_per_scope",
            "maxAppealsPerScope",
            "min_appeal_reason_length",
            "minAppealReasonLength",
            "review_timeout_blocks",
            "reviewTimeoutBlocks",
            "enabled",
            "require_escrow_deposit",
            "requireEscrowDeposit",
            "escrow_deposit_amount",
            "escrowDepositAmount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealWindowBlocks,
            MaxAppealsPerScope,
            MinAppealReasonLength,
            ReviewTimeoutBlocks,
            Enabled,
            RequireEscrowDeposit,
            EscrowDepositAmount,
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
                            "appealWindowBlocks" | "appeal_window_blocks" => Ok(GeneratedField::AppealWindowBlocks),
                            "maxAppealsPerScope" | "max_appeals_per_scope" => Ok(GeneratedField::MaxAppealsPerScope),
                            "minAppealReasonLength" | "min_appeal_reason_length" => Ok(GeneratedField::MinAppealReasonLength),
                            "reviewTimeoutBlocks" | "review_timeout_blocks" => Ok(GeneratedField::ReviewTimeoutBlocks),
                            "enabled" => Ok(GeneratedField::Enabled),
                            "requireEscrowDeposit" | "require_escrow_deposit" => Ok(GeneratedField::RequireEscrowDeposit),
                            "escrowDepositAmount" | "escrow_deposit_amount" => Ok(GeneratedField::EscrowDepositAmount),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AppealParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.AppealParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AppealParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_window_blocks__ = None;
                let mut max_appeals_per_scope__ = None;
                let mut min_appeal_reason_length__ = None;
                let mut review_timeout_blocks__ = None;
                let mut enabled__ = None;
                let mut require_escrow_deposit__ = None;
                let mut escrow_deposit_amount__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealWindowBlocks => {
                            if appeal_window_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealWindowBlocks"));
                            }
                            appeal_window_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxAppealsPerScope => {
                            if max_appeals_per_scope__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxAppealsPerScope"));
                            }
                            max_appeals_per_scope__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinAppealReasonLength => {
                            if min_appeal_reason_length__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minAppealReasonLength"));
                            }
                            min_appeal_reason_length__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReviewTimeoutBlocks => {
                            if review_timeout_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewTimeoutBlocks"));
                            }
                            review_timeout_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Enabled => {
                            if enabled__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enabled"));
                            }
                            enabled__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireEscrowDeposit => {
                            if require_escrow_deposit__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireEscrowDeposit"));
                            }
                            require_escrow_deposit__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EscrowDepositAmount => {
                            if escrow_deposit_amount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("escrowDepositAmount"));
                            }
                            escrow_deposit_amount__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AppealParams {
                    appeal_window_blocks: appeal_window_blocks__.unwrap_or_default(),
                    max_appeals_per_scope: max_appeals_per_scope__.unwrap_or_default(),
                    min_appeal_reason_length: min_appeal_reason_length__.unwrap_or_default(),
                    review_timeout_blocks: review_timeout_blocks__.unwrap_or_default(),
                    enabled: enabled__.unwrap_or_default(),
                    require_escrow_deposit: require_escrow_deposit__.unwrap_or_default(),
                    escrow_deposit_amount: escrow_deposit_amount__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.AppealParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AppealRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.original_status.is_empty() {
            len += 1;
        }
        if self.original_score != 0 {
            len += 1;
        }
        if !self.appeal_reason.is_empty() {
            len += 1;
        }
        if !self.evidence_hashes.is_empty() {
            len += 1;
        }
        if self.submitted_at != 0 {
            len += 1;
        }
        if self.submitted_at_time != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.reviewer_address.is_empty() {
            len += 1;
        }
        if self.claimed_at != 0 {
            len += 1;
        }
        if self.reviewed_at != 0 {
            len += 1;
        }
        if self.reviewed_at_time != 0 {
            len += 1;
        }
        if !self.resolution_reason.is_empty() {
            len += 1;
        }
        if self.score_adjustment != 0 {
            len += 1;
        }
        if self.appeal_number != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.AppealRecord", len)?;
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.original_status.is_empty() {
            struct_ser.serialize_field("originalStatus", &self.original_status)?;
        }
        if self.original_score != 0 {
            struct_ser.serialize_field("originalScore", &self.original_score)?;
        }
        if !self.appeal_reason.is_empty() {
            struct_ser.serialize_field("appealReason", &self.appeal_reason)?;
        }
        if !self.evidence_hashes.is_empty() {
            struct_ser.serialize_field("evidenceHashes", &self.evidence_hashes)?;
        }
        if self.submitted_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("submittedAt", ToString::to_string(&self.submitted_at).as_str())?;
        }
        if self.submitted_at_time != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("submittedAtTime", ToString::to_string(&self.submitted_at_time).as_str())?;
        }
        if self.status != 0 {
            let v = AppealStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.reviewer_address.is_empty() {
            struct_ser.serialize_field("reviewerAddress", &self.reviewer_address)?;
        }
        if self.claimed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("claimedAt", ToString::to_string(&self.claimed_at).as_str())?;
        }
        if self.reviewed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reviewedAt", ToString::to_string(&self.reviewed_at).as_str())?;
        }
        if self.reviewed_at_time != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reviewedAtTime", ToString::to_string(&self.reviewed_at_time).as_str())?;
        }
        if !self.resolution_reason.is_empty() {
            struct_ser.serialize_field("resolutionReason", &self.resolution_reason)?;
        }
        if self.score_adjustment != 0 {
            struct_ser.serialize_field("scoreAdjustment", &self.score_adjustment)?;
        }
        if self.appeal_number != 0 {
            struct_ser.serialize_field("appealNumber", &self.appeal_number)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AppealRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_id",
            "appealId",
            "account_address",
            "accountAddress",
            "scope_id",
            "scopeId",
            "original_status",
            "originalStatus",
            "original_score",
            "originalScore",
            "appeal_reason",
            "appealReason",
            "evidence_hashes",
            "evidenceHashes",
            "submitted_at",
            "submittedAt",
            "submitted_at_time",
            "submittedAtTime",
            "status",
            "reviewer_address",
            "reviewerAddress",
            "claimed_at",
            "claimedAt",
            "reviewed_at",
            "reviewedAt",
            "reviewed_at_time",
            "reviewedAtTime",
            "resolution_reason",
            "resolutionReason",
            "score_adjustment",
            "scoreAdjustment",
            "appeal_number",
            "appealNumber",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealId,
            AccountAddress,
            ScopeId,
            OriginalStatus,
            OriginalScore,
            AppealReason,
            EvidenceHashes,
            SubmittedAt,
            SubmittedAtTime,
            Status,
            ReviewerAddress,
            ClaimedAt,
            ReviewedAt,
            ReviewedAtTime,
            ResolutionReason,
            ScoreAdjustment,
            AppealNumber,
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
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "originalStatus" | "original_status" => Ok(GeneratedField::OriginalStatus),
                            "originalScore" | "original_score" => Ok(GeneratedField::OriginalScore),
                            "appealReason" | "appeal_reason" => Ok(GeneratedField::AppealReason),
                            "evidenceHashes" | "evidence_hashes" => Ok(GeneratedField::EvidenceHashes),
                            "submittedAt" | "submitted_at" => Ok(GeneratedField::SubmittedAt),
                            "submittedAtTime" | "submitted_at_time" => Ok(GeneratedField::SubmittedAtTime),
                            "status" => Ok(GeneratedField::Status),
                            "reviewerAddress" | "reviewer_address" => Ok(GeneratedField::ReviewerAddress),
                            "claimedAt" | "claimed_at" => Ok(GeneratedField::ClaimedAt),
                            "reviewedAt" | "reviewed_at" => Ok(GeneratedField::ReviewedAt),
                            "reviewedAtTime" | "reviewed_at_time" => Ok(GeneratedField::ReviewedAtTime),
                            "resolutionReason" | "resolution_reason" => Ok(GeneratedField::ResolutionReason),
                            "scoreAdjustment" | "score_adjustment" => Ok(GeneratedField::ScoreAdjustment),
                            "appealNumber" | "appeal_number" => Ok(GeneratedField::AppealNumber),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AppealRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.AppealRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AppealRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_id__ = None;
                let mut account_address__ = None;
                let mut scope_id__ = None;
                let mut original_status__ = None;
                let mut original_score__ = None;
                let mut appeal_reason__ = None;
                let mut evidence_hashes__ = None;
                let mut submitted_at__ = None;
                let mut submitted_at_time__ = None;
                let mut status__ = None;
                let mut reviewer_address__ = None;
                let mut claimed_at__ = None;
                let mut reviewed_at__ = None;
                let mut reviewed_at_time__ = None;
                let mut resolution_reason__ = None;
                let mut score_adjustment__ = None;
                let mut appeal_number__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OriginalStatus => {
                            if original_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("originalStatus"));
                            }
                            original_status__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OriginalScore => {
                            if original_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("originalScore"));
                            }
                            original_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AppealReason => {
                            if appeal_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealReason"));
                            }
                            appeal_reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EvidenceHashes => {
                            if evidence_hashes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidenceHashes"));
                            }
                            evidence_hashes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubmittedAt => {
                            if submitted_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submittedAt"));
                            }
                            submitted_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SubmittedAtTime => {
                            if submitted_at_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submittedAtTime"));
                            }
                            submitted_at_time__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<AppealStatus>()? as i32);
                        }
                        GeneratedField::ReviewerAddress => {
                            if reviewer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewerAddress"));
                            }
                            reviewer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClaimedAt => {
                            if claimed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("claimedAt"));
                            }
                            claimed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReviewedAt => {
                            if reviewed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewedAt"));
                            }
                            reviewed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReviewedAtTime => {
                            if reviewed_at_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewedAtTime"));
                            }
                            reviewed_at_time__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ResolutionReason => {
                            if resolution_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolutionReason"));
                            }
                            resolution_reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScoreAdjustment => {
                            if score_adjustment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreAdjustment"));
                            }
                            score_adjustment__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AppealNumber => {
                            if appeal_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealNumber"));
                            }
                            appeal_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AppealRecord {
                    appeal_id: appeal_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    original_status: original_status__.unwrap_or_default(),
                    original_score: original_score__.unwrap_or_default(),
                    appeal_reason: appeal_reason__.unwrap_or_default(),
                    evidence_hashes: evidence_hashes__.unwrap_or_default(),
                    submitted_at: submitted_at__.unwrap_or_default(),
                    submitted_at_time: submitted_at_time__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    reviewer_address: reviewer_address__.unwrap_or_default(),
                    claimed_at: claimed_at__.unwrap_or_default(),
                    reviewed_at: reviewed_at__.unwrap_or_default(),
                    reviewed_at_time: reviewed_at_time__.unwrap_or_default(),
                    resolution_reason: resolution_reason__.unwrap_or_default(),
                    score_adjustment: score_adjustment__.unwrap_or_default(),
                    appeal_number: appeal_number__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.AppealRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for AppealStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "APPEAL_STATUS_UNSPECIFIED",
            Self::Pending => "APPEAL_STATUS_PENDING",
            Self::Reviewing => "APPEAL_STATUS_REVIEWING",
            Self::Approved => "APPEAL_STATUS_APPROVED",
            Self::Rejected => "APPEAL_STATUS_REJECTED",
            Self::Withdrawn => "APPEAL_STATUS_WITHDRAWN",
            Self::Expired => "APPEAL_STATUS_EXPIRED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for AppealStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "APPEAL_STATUS_UNSPECIFIED",
            "APPEAL_STATUS_PENDING",
            "APPEAL_STATUS_REVIEWING",
            "APPEAL_STATUS_APPROVED",
            "APPEAL_STATUS_REJECTED",
            "APPEAL_STATUS_WITHDRAWN",
            "APPEAL_STATUS_EXPIRED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AppealStatus;

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
                    "APPEAL_STATUS_UNSPECIFIED" => Ok(AppealStatus::Unspecified),
                    "APPEAL_STATUS_PENDING" => Ok(AppealStatus::Pending),
                    "APPEAL_STATUS_REVIEWING" => Ok(AppealStatus::Reviewing),
                    "APPEAL_STATUS_APPROVED" => Ok(AppealStatus::Approved),
                    "APPEAL_STATUS_REJECTED" => Ok(AppealStatus::Rejected),
                    "APPEAL_STATUS_WITHDRAWN" => Ok(AppealStatus::Withdrawn),
                    "APPEAL_STATUS_EXPIRED" => Ok(AppealStatus::Expired),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for AppealSummary {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.total_appeals != 0 {
            len += 1;
        }
        if self.pending_appeals != 0 {
            len += 1;
        }
        if self.approved_appeals != 0 {
            len += 1;
        }
        if self.rejected_appeals != 0 {
            len += 1;
        }
        if self.withdrawn_appeals != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.AppealSummary", len)?;
        if self.total_appeals != 0 {
            struct_ser.serialize_field("totalAppeals", &self.total_appeals)?;
        }
        if self.pending_appeals != 0 {
            struct_ser.serialize_field("pendingAppeals", &self.pending_appeals)?;
        }
        if self.approved_appeals != 0 {
            struct_ser.serialize_field("approvedAppeals", &self.approved_appeals)?;
        }
        if self.rejected_appeals != 0 {
            struct_ser.serialize_field("rejectedAppeals", &self.rejected_appeals)?;
        }
        if self.withdrawn_appeals != 0 {
            struct_ser.serialize_field("withdrawnAppeals", &self.withdrawn_appeals)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for AppealSummary {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "total_appeals",
            "totalAppeals",
            "pending_appeals",
            "pendingAppeals",
            "approved_appeals",
            "approvedAppeals",
            "rejected_appeals",
            "rejectedAppeals",
            "withdrawn_appeals",
            "withdrawnAppeals",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TotalAppeals,
            PendingAppeals,
            ApprovedAppeals,
            RejectedAppeals,
            WithdrawnAppeals,
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
                            "totalAppeals" | "total_appeals" => Ok(GeneratedField::TotalAppeals),
                            "pendingAppeals" | "pending_appeals" => Ok(GeneratedField::PendingAppeals),
                            "approvedAppeals" | "approved_appeals" => Ok(GeneratedField::ApprovedAppeals),
                            "rejectedAppeals" | "rejected_appeals" => Ok(GeneratedField::RejectedAppeals),
                            "withdrawnAppeals" | "withdrawn_appeals" => Ok(GeneratedField::WithdrawnAppeals),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = AppealSummary;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.AppealSummary")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<AppealSummary, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut total_appeals__ = None;
                let mut pending_appeals__ = None;
                let mut approved_appeals__ = None;
                let mut rejected_appeals__ = None;
                let mut withdrawn_appeals__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TotalAppeals => {
                            if total_appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalAppeals"));
                            }
                            total_appeals__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PendingAppeals => {
                            if pending_appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pendingAppeals"));
                            }
                            pending_appeals__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ApprovedAppeals => {
                            if approved_appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("approvedAppeals"));
                            }
                            approved_appeals__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RejectedAppeals => {
                            if rejected_appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rejectedAppeals"));
                            }
                            rejected_appeals__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::WithdrawnAppeals => {
                            if withdrawn_appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("withdrawnAppeals"));
                            }
                            withdrawn_appeals__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(AppealSummary {
                    total_appeals: total_appeals__.unwrap_or_default(),
                    pending_appeals: pending_appeals__.unwrap_or_default(),
                    approved_appeals: approved_appeals__.unwrap_or_default(),
                    rejected_appeals: rejected_appeals__.unwrap_or_default(),
                    withdrawn_appeals: withdrawn_appeals__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.AppealSummary", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ApprovedClient {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.client_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.public_key.is_empty() {
            len += 1;
        }
        if self.active {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.deactivated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ApprovedClient", len)?;
        if !self.client_id.is_empty() {
            struct_ser.serialize_field("clientId", &self.client_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.public_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("publicKey", pbjson::private::base64::encode(&self.public_key).as_str())?;
        }
        if self.active {
            struct_ser.serialize_field("active", &self.active)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.deactivated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("deactivatedAt", ToString::to_string(&self.deactivated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ApprovedClient {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "client_id",
            "clientId",
            "name",
            "public_key",
            "publicKey",
            "active",
            "created_at",
            "createdAt",
            "deactivated_at",
            "deactivatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ClientId,
            Name,
            PublicKey,
            Active,
            CreatedAt,
            DeactivatedAt,
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
                            "clientId" | "client_id" => Ok(GeneratedField::ClientId),
                            "name" => Ok(GeneratedField::Name),
                            "publicKey" | "public_key" => Ok(GeneratedField::PublicKey),
                            "active" => Ok(GeneratedField::Active),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "deactivatedAt" | "deactivated_at" => Ok(GeneratedField::DeactivatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ApprovedClient;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ApprovedClient")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ApprovedClient, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut client_id__ = None;
                let mut name__ = None;
                let mut public_key__ = None;
                let mut active__ = None;
                let mut created_at__ = None;
                let mut deactivated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ClientId => {
                            if client_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientId"));
                            }
                            client_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PublicKey => {
                            if public_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("publicKey"));
                            }
                            public_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Active => {
                            if active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("active"));
                            }
                            active__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DeactivatedAt => {
                            if deactivated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deactivatedAt"));
                            }
                            deactivated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ApprovedClient {
                    client_id: client_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    public_key: public_key__.unwrap_or_default(),
                    active: active__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    deactivated_at: deactivated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ApprovedClient", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for BorderlineParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.lower_threshold != 0 {
            len += 1;
        }
        if self.upper_threshold != 0 {
            len += 1;
        }
        if self.mfa_timeout_blocks != 0 {
            len += 1;
        }
        if self.required_factors != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.BorderlineParams", len)?;
        if self.lower_threshold != 0 {
            struct_ser.serialize_field("lowerThreshold", &self.lower_threshold)?;
        }
        if self.upper_threshold != 0 {
            struct_ser.serialize_field("upperThreshold", &self.upper_threshold)?;
        }
        if self.mfa_timeout_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("mfaTimeoutBlocks", ToString::to_string(&self.mfa_timeout_blocks).as_str())?;
        }
        if self.required_factors != 0 {
            struct_ser.serialize_field("requiredFactors", &self.required_factors)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BorderlineParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "lower_threshold",
            "lowerThreshold",
            "upper_threshold",
            "upperThreshold",
            "mfa_timeout_blocks",
            "mfaTimeoutBlocks",
            "required_factors",
            "requiredFactors",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            LowerThreshold,
            UpperThreshold,
            MfaTimeoutBlocks,
            RequiredFactors,
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
                            "lowerThreshold" | "lower_threshold" => Ok(GeneratedField::LowerThreshold),
                            "upperThreshold" | "upper_threshold" => Ok(GeneratedField::UpperThreshold),
                            "mfaTimeoutBlocks" | "mfa_timeout_blocks" => Ok(GeneratedField::MfaTimeoutBlocks),
                            "requiredFactors" | "required_factors" => Ok(GeneratedField::RequiredFactors),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BorderlineParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.BorderlineParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<BorderlineParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut lower_threshold__ = None;
                let mut upper_threshold__ = None;
                let mut mfa_timeout_blocks__ = None;
                let mut required_factors__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::LowerThreshold => {
                            if lower_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lowerThreshold"));
                            }
                            lower_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpperThreshold => {
                            if upper_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("upperThreshold"));
                            }
                            upper_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MfaTimeoutBlocks => {
                            if mfa_timeout_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaTimeoutBlocks"));
                            }
                            mfa_timeout_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequiredFactors => {
                            if required_factors__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredFactors"));
                            }
                            required_factors__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(BorderlineParams {
                    lower_threshold: lower_threshold__.unwrap_or_default(),
                    upper_threshold: upper_threshold__.unwrap_or_default(),
                    mfa_timeout_blocks: mfa_timeout_blocks__.unwrap_or_default(),
                    required_factors: required_factors__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.BorderlineParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceAttestation {
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
        if self.attested_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if !self.attestation_type.is_empty() {
            len += 1;
        }
        if !self.attestation_hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ComplianceAttestation", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if self.attested_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("attestedAt", ToString::to_string(&self.attested_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if !self.attestation_type.is_empty() {
            struct_ser.serialize_field("attestationType", &self.attestation_type)?;
        }
        if !self.attestation_hash.is_empty() {
            struct_ser.serialize_field("attestationHash", &self.attestation_hash)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceAttestation {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "attested_at",
            "attestedAt",
            "expires_at",
            "expiresAt",
            "attestation_type",
            "attestationType",
            "attestation_hash",
            "attestationHash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            AttestedAt,
            ExpiresAt,
            AttestationType,
            AttestationHash,
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
                            "attestedAt" | "attested_at" => Ok(GeneratedField::AttestedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "attestationType" | "attestation_type" => Ok(GeneratedField::AttestationType),
                            "attestationHash" | "attestation_hash" => Ok(GeneratedField::AttestationHash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceAttestation;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ComplianceAttestation")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ComplianceAttestation, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut attested_at__ = None;
                let mut expires_at__ = None;
                let mut attestation_type__ = None;
                let mut attestation_hash__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AttestedAt => {
                            if attested_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestedAt"));
                            }
                            attested_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AttestationType => {
                            if attestation_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationType"));
                            }
                            attestation_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AttestationHash => {
                            if attestation_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationHash"));
                            }
                            attestation_hash__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ComplianceAttestation {
                    validator_address: validator_address__.unwrap_or_default(),
                    attested_at: attested_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    attestation_type: attestation_type__.unwrap_or_default(),
                    attestation_hash: attestation_hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ComplianceAttestation", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceCheckResult {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.check_type != 0 {
            len += 1;
        }
        if self.passed {
            len += 1;
        }
        if !self.details.is_empty() {
            len += 1;
        }
        if self.match_score != 0 {
            len += 1;
        }
        if self.checked_at != 0 {
            len += 1;
        }
        if !self.provider_id.is_empty() {
            len += 1;
        }
        if !self.reference_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ComplianceCheckResult", len)?;
        if self.check_type != 0 {
            let v = ComplianceCheckType::try_from(self.check_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.check_type)))?;
            struct_ser.serialize_field("checkType", &v)?;
        }
        if self.passed {
            struct_ser.serialize_field("passed", &self.passed)?;
        }
        if !self.details.is_empty() {
            struct_ser.serialize_field("details", &self.details)?;
        }
        if self.match_score != 0 {
            struct_ser.serialize_field("matchScore", &self.match_score)?;
        }
        if self.checked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("checkedAt", ToString::to_string(&self.checked_at).as_str())?;
        }
        if !self.provider_id.is_empty() {
            struct_ser.serialize_field("providerId", &self.provider_id)?;
        }
        if !self.reference_id.is_empty() {
            struct_ser.serialize_field("referenceId", &self.reference_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceCheckResult {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "check_type",
            "checkType",
            "passed",
            "details",
            "match_score",
            "matchScore",
            "checked_at",
            "checkedAt",
            "provider_id",
            "providerId",
            "reference_id",
            "referenceId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            CheckType,
            Passed,
            Details,
            MatchScore,
            CheckedAt,
            ProviderId,
            ReferenceId,
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
                            "checkType" | "check_type" => Ok(GeneratedField::CheckType),
                            "passed" => Ok(GeneratedField::Passed),
                            "details" => Ok(GeneratedField::Details),
                            "matchScore" | "match_score" => Ok(GeneratedField::MatchScore),
                            "checkedAt" | "checked_at" => Ok(GeneratedField::CheckedAt),
                            "providerId" | "provider_id" => Ok(GeneratedField::ProviderId),
                            "referenceId" | "reference_id" => Ok(GeneratedField::ReferenceId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceCheckResult;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ComplianceCheckResult")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ComplianceCheckResult, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut check_type__ = None;
                let mut passed__ = None;
                let mut details__ = None;
                let mut match_score__ = None;
                let mut checked_at__ = None;
                let mut provider_id__ = None;
                let mut reference_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::CheckType => {
                            if check_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("checkType"));
                            }
                            check_type__ = Some(map_.next_value::<ComplianceCheckType>()? as i32);
                        }
                        GeneratedField::Passed => {
                            if passed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("passed"));
                            }
                            passed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Details => {
                            if details__.is_some() {
                                return Err(serde::de::Error::duplicate_field("details"));
                            }
                            details__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MatchScore => {
                            if match_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("matchScore"));
                            }
                            match_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CheckedAt => {
                            if checked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("checkedAt"));
                            }
                            checked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ProviderId => {
                            if provider_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerId"));
                            }
                            provider_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReferenceId => {
                            if reference_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("referenceId"));
                            }
                            reference_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ComplianceCheckResult {
                    check_type: check_type__.unwrap_or_default(),
                    passed: passed__.unwrap_or_default(),
                    details: details__.unwrap_or_default(),
                    match_score: match_score__.unwrap_or_default(),
                    checked_at: checked_at__.unwrap_or_default(),
                    provider_id: provider_id__.unwrap_or_default(),
                    reference_id: reference_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ComplianceCheckResult", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceCheckType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::ComplianceCheckSanctionList => "COMPLIANCE_CHECK_SANCTION_LIST",
            Self::ComplianceCheckPep => "COMPLIANCE_CHECK_PEP",
            Self::ComplianceCheckAdverseMedia => "COMPLIANCE_CHECK_ADVERSE_MEDIA",
            Self::ComplianceCheckGeographic => "COMPLIANCE_CHECK_GEOGRAPHIC",
            Self::ComplianceCheckWatchlist => "COMPLIANCE_CHECK_WATCHLIST",
            Self::ComplianceCheckDocumentVerification => "COMPLIANCE_CHECK_DOCUMENT_VERIFICATION",
            Self::ComplianceCheckAmlRisk => "COMPLIANCE_CHECK_AML_RISK",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceCheckType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "COMPLIANCE_CHECK_SANCTION_LIST",
            "COMPLIANCE_CHECK_PEP",
            "COMPLIANCE_CHECK_ADVERSE_MEDIA",
            "COMPLIANCE_CHECK_GEOGRAPHIC",
            "COMPLIANCE_CHECK_WATCHLIST",
            "COMPLIANCE_CHECK_DOCUMENT_VERIFICATION",
            "COMPLIANCE_CHECK_AML_RISK",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceCheckType;

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
                    "COMPLIANCE_CHECK_SANCTION_LIST" => Ok(ComplianceCheckType::ComplianceCheckSanctionList),
                    "COMPLIANCE_CHECK_PEP" => Ok(ComplianceCheckType::ComplianceCheckPep),
                    "COMPLIANCE_CHECK_ADVERSE_MEDIA" => Ok(ComplianceCheckType::ComplianceCheckAdverseMedia),
                    "COMPLIANCE_CHECK_GEOGRAPHIC" => Ok(ComplianceCheckType::ComplianceCheckGeographic),
                    "COMPLIANCE_CHECK_WATCHLIST" => Ok(ComplianceCheckType::ComplianceCheckWatchlist),
                    "COMPLIANCE_CHECK_DOCUMENT_VERIFICATION" => Ok(ComplianceCheckType::ComplianceCheckDocumentVerification),
                    "COMPLIANCE_CHECK_AML_RISK" => Ok(ComplianceCheckType::ComplianceCheckAmlRisk),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.require_sanction_check {
            len += 1;
        }
        if self.require_pep_check {
            len += 1;
        }
        if self.check_expiry_blocks != 0 {
            len += 1;
        }
        if self.risk_score_threshold != 0 {
            len += 1;
        }
        if !self.restricted_countries.is_empty() {
            len += 1;
        }
        if self.min_attestations_required != 0 {
            len += 1;
        }
        if self.enable_auto_expiry {
            len += 1;
        }
        if self.require_document_verification {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ComplianceParams", len)?;
        if self.require_sanction_check {
            struct_ser.serialize_field("requireSanctionCheck", &self.require_sanction_check)?;
        }
        if self.require_pep_check {
            struct_ser.serialize_field("requirePepCheck", &self.require_pep_check)?;
        }
        if self.check_expiry_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("checkExpiryBlocks", ToString::to_string(&self.check_expiry_blocks).as_str())?;
        }
        if self.risk_score_threshold != 0 {
            struct_ser.serialize_field("riskScoreThreshold", &self.risk_score_threshold)?;
        }
        if !self.restricted_countries.is_empty() {
            struct_ser.serialize_field("restrictedCountries", &self.restricted_countries)?;
        }
        if self.min_attestations_required != 0 {
            struct_ser.serialize_field("minAttestationsRequired", &self.min_attestations_required)?;
        }
        if self.enable_auto_expiry {
            struct_ser.serialize_field("enableAutoExpiry", &self.enable_auto_expiry)?;
        }
        if self.require_document_verification {
            struct_ser.serialize_field("requireDocumentVerification", &self.require_document_verification)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "require_sanction_check",
            "requireSanctionCheck",
            "require_pep_check",
            "requirePepCheck",
            "check_expiry_blocks",
            "checkExpiryBlocks",
            "risk_score_threshold",
            "riskScoreThreshold",
            "restricted_countries",
            "restrictedCountries",
            "min_attestations_required",
            "minAttestationsRequired",
            "enable_auto_expiry",
            "enableAutoExpiry",
            "require_document_verification",
            "requireDocumentVerification",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RequireSanctionCheck,
            RequirePepCheck,
            CheckExpiryBlocks,
            RiskScoreThreshold,
            RestrictedCountries,
            MinAttestationsRequired,
            EnableAutoExpiry,
            RequireDocumentVerification,
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
                            "requireSanctionCheck" | "require_sanction_check" => Ok(GeneratedField::RequireSanctionCheck),
                            "requirePepCheck" | "require_pep_check" => Ok(GeneratedField::RequirePepCheck),
                            "checkExpiryBlocks" | "check_expiry_blocks" => Ok(GeneratedField::CheckExpiryBlocks),
                            "riskScoreThreshold" | "risk_score_threshold" => Ok(GeneratedField::RiskScoreThreshold),
                            "restrictedCountries" | "restricted_countries" => Ok(GeneratedField::RestrictedCountries),
                            "minAttestationsRequired" | "min_attestations_required" => Ok(GeneratedField::MinAttestationsRequired),
                            "enableAutoExpiry" | "enable_auto_expiry" => Ok(GeneratedField::EnableAutoExpiry),
                            "requireDocumentVerification" | "require_document_verification" => Ok(GeneratedField::RequireDocumentVerification),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ComplianceParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ComplianceParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut require_sanction_check__ = None;
                let mut require_pep_check__ = None;
                let mut check_expiry_blocks__ = None;
                let mut risk_score_threshold__ = None;
                let mut restricted_countries__ = None;
                let mut min_attestations_required__ = None;
                let mut enable_auto_expiry__ = None;
                let mut require_document_verification__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RequireSanctionCheck => {
                            if require_sanction_check__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireSanctionCheck"));
                            }
                            require_sanction_check__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequirePepCheck => {
                            if require_pep_check__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requirePepCheck"));
                            }
                            require_pep_check__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CheckExpiryBlocks => {
                            if check_expiry_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("checkExpiryBlocks"));
                            }
                            check_expiry_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RiskScoreThreshold => {
                            if risk_score_threshold__.is_some() {
                                return Err(serde::de::Error::duplicate_field("riskScoreThreshold"));
                            }
                            risk_score_threshold__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RestrictedCountries => {
                            if restricted_countries__.is_some() {
                                return Err(serde::de::Error::duplicate_field("restrictedCountries"));
                            }
                            restricted_countries__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MinAttestationsRequired => {
                            if min_attestations_required__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minAttestationsRequired"));
                            }
                            min_attestations_required__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EnableAutoExpiry => {
                            if enable_auto_expiry__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enableAutoExpiry"));
                            }
                            enable_auto_expiry__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireDocumentVerification => {
                            if require_document_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireDocumentVerification"));
                            }
                            require_document_verification__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ComplianceParams {
                    require_sanction_check: require_sanction_check__.unwrap_or_default(),
                    require_pep_check: require_pep_check__.unwrap_or_default(),
                    check_expiry_blocks: check_expiry_blocks__.unwrap_or_default(),
                    risk_score_threshold: risk_score_threshold__.unwrap_or_default(),
                    restricted_countries: restricted_countries__.unwrap_or_default(),
                    min_attestations_required: min_attestations_required__.unwrap_or_default(),
                    enable_auto_expiry: enable_auto_expiry__.unwrap_or_default(),
                    require_document_verification: require_document_verification__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ComplianceParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceProvider {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.provider_address.is_empty() {
            len += 1;
        }
        if !self.supported_check_types.is_empty() {
            len += 1;
        }
        if self.is_active {
            len += 1;
        }
        if self.registered_at != 0 {
            len += 1;
        }
        if self.last_active_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ComplianceProvider", len)?;
        if !self.provider_id.is_empty() {
            struct_ser.serialize_field("providerId", &self.provider_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.supported_check_types.is_empty() {
            let v = self.supported_check_types.iter().cloned().map(|v| {
                ComplianceCheckType::try_from(v)
                    .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", v)))
                }).collect::<std::result::Result<Vec<_>, _>>()?;
            struct_ser.serialize_field("supportedCheckTypes", &v)?;
        }
        if self.is_active {
            struct_ser.serialize_field("isActive", &self.is_active)?;
        }
        if self.registered_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("registeredAt", ToString::to_string(&self.registered_at).as_str())?;
        }
        if self.last_active_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastActiveAt", ToString::to_string(&self.last_active_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceProvider {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_id",
            "providerId",
            "name",
            "provider_address",
            "providerAddress",
            "supported_check_types",
            "supportedCheckTypes",
            "is_active",
            "isActive",
            "registered_at",
            "registeredAt",
            "last_active_at",
            "lastActiveAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderId,
            Name,
            ProviderAddress,
            SupportedCheckTypes,
            IsActive,
            RegisteredAt,
            LastActiveAt,
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
                            "providerId" | "provider_id" => Ok(GeneratedField::ProviderId),
                            "name" => Ok(GeneratedField::Name),
                            "providerAddress" | "provider_address" => Ok(GeneratedField::ProviderAddress),
                            "supportedCheckTypes" | "supported_check_types" => Ok(GeneratedField::SupportedCheckTypes),
                            "isActive" | "is_active" => Ok(GeneratedField::IsActive),
                            "registeredAt" | "registered_at" => Ok(GeneratedField::RegisteredAt),
                            "lastActiveAt" | "last_active_at" => Ok(GeneratedField::LastActiveAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceProvider;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ComplianceProvider")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ComplianceProvider, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_id__ = None;
                let mut name__ = None;
                let mut provider_address__ = None;
                let mut supported_check_types__ = None;
                let mut is_active__ = None;
                let mut registered_at__ = None;
                let mut last_active_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderId => {
                            if provider_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerId"));
                            }
                            provider_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SupportedCheckTypes => {
                            if supported_check_types__.is_some() {
                                return Err(serde::de::Error::duplicate_field("supportedCheckTypes"));
                            }
                            supported_check_types__ = Some(map_.next_value::<Vec<ComplianceCheckType>>()?.into_iter().map(|x| x as i32).collect());
                        }
                        GeneratedField::IsActive => {
                            if is_active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isActive"));
                            }
                            is_active__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RegisteredAt => {
                            if registered_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registeredAt"));
                            }
                            registered_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastActiveAt => {
                            if last_active_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastActiveAt"));
                            }
                            last_active_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ComplianceProvider {
                    provider_id: provider_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    provider_address: provider_address__.unwrap_or_default(),
                    supported_check_types: supported_check_types__.unwrap_or_default(),
                    is_active: is_active__.unwrap_or_default(),
                    registered_at: registered_at__.unwrap_or_default(),
                    last_active_at: last_active_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ComplianceProvider", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.check_results.is_empty() {
            len += 1;
        }
        if self.last_checked_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if self.risk_score != 0 {
            len += 1;
        }
        if !self.restricted_regions.is_empty() {
            len += 1;
        }
        if !self.attestations.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        if !self.notes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ComplianceRecord", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.status != 0 {
            let v = ComplianceStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.check_results.is_empty() {
            struct_ser.serialize_field("checkResults", &self.check_results)?;
        }
        if self.last_checked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastCheckedAt", ToString::to_string(&self.last_checked_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if self.risk_score != 0 {
            struct_ser.serialize_field("riskScore", &self.risk_score)?;
        }
        if !self.restricted_regions.is_empty() {
            struct_ser.serialize_field("restrictedRegions", &self.restricted_regions)?;
        }
        if !self.attestations.is_empty() {
            struct_ser.serialize_field("attestations", &self.attestations)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        if !self.notes.is_empty() {
            struct_ser.serialize_field("notes", &self.notes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "status",
            "check_results",
            "checkResults",
            "last_checked_at",
            "lastCheckedAt",
            "expires_at",
            "expiresAt",
            "risk_score",
            "riskScore",
            "restricted_regions",
            "restrictedRegions",
            "attestations",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "notes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            Status,
            CheckResults,
            LastCheckedAt,
            ExpiresAt,
            RiskScore,
            RestrictedRegions,
            Attestations,
            CreatedAt,
            UpdatedAt,
            Notes,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "status" => Ok(GeneratedField::Status),
                            "checkResults" | "check_results" => Ok(GeneratedField::CheckResults),
                            "lastCheckedAt" | "last_checked_at" => Ok(GeneratedField::LastCheckedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "riskScore" | "risk_score" => Ok(GeneratedField::RiskScore),
                            "restrictedRegions" | "restricted_regions" => Ok(GeneratedField::RestrictedRegions),
                            "attestations" => Ok(GeneratedField::Attestations),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "notes" => Ok(GeneratedField::Notes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ComplianceRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ComplianceRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut status__ = None;
                let mut check_results__ = None;
                let mut last_checked_at__ = None;
                let mut expires_at__ = None;
                let mut risk_score__ = None;
                let mut restricted_regions__ = None;
                let mut attestations__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut notes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ComplianceStatus>()? as i32);
                        }
                        GeneratedField::CheckResults => {
                            if check_results__.is_some() {
                                return Err(serde::de::Error::duplicate_field("checkResults"));
                            }
                            check_results__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastCheckedAt => {
                            if last_checked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastCheckedAt"));
                            }
                            last_checked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RiskScore => {
                            if risk_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("riskScore"));
                            }
                            risk_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RestrictedRegions => {
                            if restricted_regions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("restrictedRegions"));
                            }
                            restricted_regions__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Attestations => {
                            if attestations__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestations"));
                            }
                            attestations__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Notes => {
                            if notes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("notes"));
                            }
                            notes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ComplianceRecord {
                    account_address: account_address__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    check_results: check_results__.unwrap_or_default(),
                    last_checked_at: last_checked_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    risk_score: risk_score__.unwrap_or_default(),
                    restricted_regions: restricted_regions__.unwrap_or_default(),
                    attestations: attestations__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                    notes: notes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ComplianceRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ComplianceStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unknown => "COMPLIANCE_STATUS_UNKNOWN",
            Self::Pending => "COMPLIANCE_STATUS_PENDING",
            Self::Cleared => "COMPLIANCE_STATUS_CLEARED",
            Self::Flagged => "COMPLIANCE_STATUS_FLAGGED",
            Self::Blocked => "COMPLIANCE_STATUS_BLOCKED",
            Self::Expired => "COMPLIANCE_STATUS_EXPIRED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ComplianceStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "COMPLIANCE_STATUS_UNKNOWN",
            "COMPLIANCE_STATUS_PENDING",
            "COMPLIANCE_STATUS_CLEARED",
            "COMPLIANCE_STATUS_FLAGGED",
            "COMPLIANCE_STATUS_BLOCKED",
            "COMPLIANCE_STATUS_EXPIRED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ComplianceStatus;

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
                    "COMPLIANCE_STATUS_UNKNOWN" => Ok(ComplianceStatus::Unknown),
                    "COMPLIANCE_STATUS_PENDING" => Ok(ComplianceStatus::Pending),
                    "COMPLIANCE_STATUS_CLEARED" => Ok(ComplianceStatus::Cleared),
                    "COMPLIANCE_STATUS_FLAGGED" => Ok(ComplianceStatus::Flagged),
                    "COMPLIANCE_STATUS_BLOCKED" => Ok(ComplianceStatus::Blocked),
                    "COMPLIANCE_STATUS_EXPIRED" => Ok(ComplianceStatus::Expired),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ConsentSettings {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.share_with_providers {
            len += 1;
        }
        if self.share_for_verification {
            len += 1;
        }
        if self.allow_re_verification {
            len += 1;
        }
        if self.allow_derived_feature_sharing {
            len += 1;
        }
        if self.consent_version != 0 {
            len += 1;
        }
        if self.last_updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ConsentSettings", len)?;
        if self.share_with_providers {
            struct_ser.serialize_field("shareWithProviders", &self.share_with_providers)?;
        }
        if self.share_for_verification {
            struct_ser.serialize_field("shareForVerification", &self.share_for_verification)?;
        }
        if self.allow_re_verification {
            struct_ser.serialize_field("allowReVerification", &self.allow_re_verification)?;
        }
        if self.allow_derived_feature_sharing {
            struct_ser.serialize_field("allowDerivedFeatureSharing", &self.allow_derived_feature_sharing)?;
        }
        if self.consent_version != 0 {
            struct_ser.serialize_field("consentVersion", &self.consent_version)?;
        }
        if self.last_updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdatedAt", ToString::to_string(&self.last_updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ConsentSettings {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "share_with_providers",
            "shareWithProviders",
            "share_for_verification",
            "shareForVerification",
            "allow_re_verification",
            "allowReVerification",
            "allow_derived_feature_sharing",
            "allowDerivedFeatureSharing",
            "consent_version",
            "consentVersion",
            "last_updated_at",
            "lastUpdatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ShareWithProviders,
            ShareForVerification,
            AllowReVerification,
            AllowDerivedFeatureSharing,
            ConsentVersion,
            LastUpdatedAt,
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
                            "shareWithProviders" | "share_with_providers" => Ok(GeneratedField::ShareWithProviders),
                            "shareForVerification" | "share_for_verification" => Ok(GeneratedField::ShareForVerification),
                            "allowReVerification" | "allow_re_verification" => Ok(GeneratedField::AllowReVerification),
                            "allowDerivedFeatureSharing" | "allow_derived_feature_sharing" => Ok(GeneratedField::AllowDerivedFeatureSharing),
                            "consentVersion" | "consent_version" => Ok(GeneratedField::ConsentVersion),
                            "lastUpdatedAt" | "last_updated_at" => Ok(GeneratedField::LastUpdatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ConsentSettings;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ConsentSettings")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ConsentSettings, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut share_with_providers__ = None;
                let mut share_for_verification__ = None;
                let mut allow_re_verification__ = None;
                let mut allow_derived_feature_sharing__ = None;
                let mut consent_version__ = None;
                let mut last_updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ShareWithProviders => {
                            if share_with_providers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("shareWithProviders"));
                            }
                            share_with_providers__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ShareForVerification => {
                            if share_for_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("shareForVerification"));
                            }
                            share_for_verification__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowReVerification => {
                            if allow_re_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowReVerification"));
                            }
                            allow_re_verification__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowDerivedFeatureSharing => {
                            if allow_derived_feature_sharing__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowDerivedFeatureSharing"));
                            }
                            allow_derived_feature_sharing__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ConsentVersion => {
                            if consent_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentVersion"));
                            }
                            consent_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastUpdatedAt => {
                            if last_updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdatedAt"));
                            }
                            last_updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ConsentSettings {
                    share_with_providers: share_with_providers__.unwrap_or_default(),
                    share_for_verification: share_for_verification__.unwrap_or_default(),
                    allow_re_verification: allow_re_verification__.unwrap_or_default(),
                    allow_derived_feature_sharing: allow_derived_feature_sharing__.unwrap_or_default(),
                    consent_version: consent_version__.unwrap_or_default(),
                    last_updated_at: last_updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ConsentSettings", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for DerivedFeatures {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.face_embedding_hash.is_empty() {
            len += 1;
        }
        if !self.doc_field_hashes.is_empty() {
            len += 1;
        }
        if !self.biometric_hash.is_empty() {
            len += 1;
        }
        if !self.liveness_proof_hash.is_empty() {
            len += 1;
        }
        if self.last_computed_at != 0 {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        if !self.computed_by.is_empty() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.feature_version != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.DerivedFeatures", len)?;
        if !self.face_embedding_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("faceEmbeddingHash", pbjson::private::base64::encode(&self.face_embedding_hash).as_str())?;
        }
        if !self.doc_field_hashes.is_empty() {
            let v: std::collections::HashMap<_, _> = self.doc_field_hashes.iter()
                .map(|(k, v)| (k, pbjson::private::base64::encode(v))).collect();
            struct_ser.serialize_field("docFieldHashes", &v)?;
        }
        if !self.biometric_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("biometricHash", pbjson::private::base64::encode(&self.biometric_hash).as_str())?;
        }
        if !self.liveness_proof_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("livenessProofHash", pbjson::private::base64::encode(&self.liveness_proof_hash).as_str())?;
        }
        if self.last_computed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastComputedAt", ToString::to_string(&self.last_computed_at).as_str())?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        if !self.computed_by.is_empty() {
            struct_ser.serialize_field("computedBy", &self.computed_by)?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if self.feature_version != 0 {
            struct_ser.serialize_field("featureVersion", &self.feature_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DerivedFeatures {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "face_embedding_hash",
            "faceEmbeddingHash",
            "doc_field_hashes",
            "docFieldHashes",
            "biometric_hash",
            "biometricHash",
            "liveness_proof_hash",
            "livenessProofHash",
            "last_computed_at",
            "lastComputedAt",
            "model_version",
            "modelVersion",
            "computed_by",
            "computedBy",
            "block_height",
            "blockHeight",
            "feature_version",
            "featureVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            FaceEmbeddingHash,
            DocFieldHashes,
            BiometricHash,
            LivenessProofHash,
            LastComputedAt,
            ModelVersion,
            ComputedBy,
            BlockHeight,
            FeatureVersion,
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
                            "faceEmbeddingHash" | "face_embedding_hash" => Ok(GeneratedField::FaceEmbeddingHash),
                            "docFieldHashes" | "doc_field_hashes" => Ok(GeneratedField::DocFieldHashes),
                            "biometricHash" | "biometric_hash" => Ok(GeneratedField::BiometricHash),
                            "livenessProofHash" | "liveness_proof_hash" => Ok(GeneratedField::LivenessProofHash),
                            "lastComputedAt" | "last_computed_at" => Ok(GeneratedField::LastComputedAt),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            "computedBy" | "computed_by" => Ok(GeneratedField::ComputedBy),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "featureVersion" | "feature_version" => Ok(GeneratedField::FeatureVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DerivedFeatures;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.DerivedFeatures")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<DerivedFeatures, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut face_embedding_hash__ = None;
                let mut doc_field_hashes__ = None;
                let mut biometric_hash__ = None;
                let mut liveness_proof_hash__ = None;
                let mut last_computed_at__ = None;
                let mut model_version__ = None;
                let mut computed_by__ = None;
                let mut block_height__ = None;
                let mut feature_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::FaceEmbeddingHash => {
                            if face_embedding_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("faceEmbeddingHash"));
                            }
                            face_embedding_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DocFieldHashes => {
                            if doc_field_hashes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("docFieldHashes"));
                            }
                            doc_field_hashes__ = Some(
                                map_.next_value::<std::collections::HashMap<_, ::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|(k,v)| (k, v.0)).collect()
                            );
                        }
                        GeneratedField::BiometricHash => {
                            if biometric_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("biometricHash"));
                            }
                            biometric_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LivenessProofHash => {
                            if liveness_proof_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("livenessProofHash"));
                            }
                            liveness_proof_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastComputedAt => {
                            if last_computed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastComputedAt"));
                            }
                            last_computed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ComputedBy => {
                            if computed_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("computedBy"));
                            }
                            computed_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::FeatureVersion => {
                            if feature_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("featureVersion"));
                            }
                            feature_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(DerivedFeatures {
                    face_embedding_hash: face_embedding_hash__.unwrap_or_default(),
                    doc_field_hashes: doc_field_hashes__.unwrap_or_default(),
                    biometric_hash: biometric_hash__.unwrap_or_default(),
                    liveness_proof_hash: liveness_proof_hash__.unwrap_or_default(),
                    last_computed_at: last_computed_at__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                    computed_by: computed_by__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    feature_version: feature_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.DerivedFeatures", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EncryptedPayloadEnvelope {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.version != 0 {
            len += 1;
        }
        if !self.algorithm_id.is_empty() {
            len += 1;
        }
        if self.algorithm_version != 0 {
            len += 1;
        }
        if !self.recipient_key_ids.is_empty() {
            len += 1;
        }
        if !self.recipient_public_keys.is_empty() {
            len += 1;
        }
        if !self.encrypted_keys.is_empty() {
            len += 1;
        }
        if !self.nonce.is_empty() {
            len += 1;
        }
        if !self.ciphertext.is_empty() {
            len += 1;
        }
        if !self.sender_signature.is_empty() {
            len += 1;
        }
        if !self.sender_pub_key.is_empty() {
            len += 1;
        }
        if !self.metadata.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.EncryptedPayloadEnvelope", len)?;
        if self.version != 0 {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if !self.algorithm_id.is_empty() {
            struct_ser.serialize_field("algorithmId", &self.algorithm_id)?;
        }
        if self.algorithm_version != 0 {
            struct_ser.serialize_field("algorithmVersion", &self.algorithm_version)?;
        }
        if !self.recipient_key_ids.is_empty() {
            struct_ser.serialize_field("recipientKeyIds", &self.recipient_key_ids)?;
        }
        if !self.recipient_public_keys.is_empty() {
            struct_ser.serialize_field("recipientPublicKeys", &self.recipient_public_keys.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.encrypted_keys.is_empty() {
            struct_ser.serialize_field("encryptedKeys", &self.encrypted_keys.iter().map(pbjson::private::base64::encode).collect::<Vec<_>>())?;
        }
        if !self.nonce.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("nonce", pbjson::private::base64::encode(&self.nonce).as_str())?;
        }
        if !self.ciphertext.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("ciphertext", pbjson::private::base64::encode(&self.ciphertext).as_str())?;
        }
        if !self.sender_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("senderSignature", pbjson::private::base64::encode(&self.sender_signature).as_str())?;
        }
        if !self.sender_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("senderPubKey", pbjson::private::base64::encode(&self.sender_pub_key).as_str())?;
        }
        if !self.metadata.is_empty() {
            struct_ser.serialize_field("metadata", &self.metadata)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EncryptedPayloadEnvelope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "version",
            "algorithm_id",
            "algorithmId",
            "algorithm_version",
            "algorithmVersion",
            "recipient_key_ids",
            "recipientKeyIds",
            "recipient_public_keys",
            "recipientPublicKeys",
            "encrypted_keys",
            "encryptedKeys",
            "nonce",
            "ciphertext",
            "sender_signature",
            "senderSignature",
            "sender_pub_key",
            "senderPubKey",
            "metadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Version,
            AlgorithmId,
            AlgorithmVersion,
            RecipientKeyIds,
            RecipientPublicKeys,
            EncryptedKeys,
            Nonce,
            Ciphertext,
            SenderSignature,
            SenderPubKey,
            Metadata,
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
                            "version" => Ok(GeneratedField::Version),
                            "algorithmId" | "algorithm_id" => Ok(GeneratedField::AlgorithmId),
                            "algorithmVersion" | "algorithm_version" => Ok(GeneratedField::AlgorithmVersion),
                            "recipientKeyIds" | "recipient_key_ids" => Ok(GeneratedField::RecipientKeyIds),
                            "recipientPublicKeys" | "recipient_public_keys" => Ok(GeneratedField::RecipientPublicKeys),
                            "encryptedKeys" | "encrypted_keys" => Ok(GeneratedField::EncryptedKeys),
                            "nonce" => Ok(GeneratedField::Nonce),
                            "ciphertext" => Ok(GeneratedField::Ciphertext),
                            "senderSignature" | "sender_signature" => Ok(GeneratedField::SenderSignature),
                            "senderPubKey" | "sender_pub_key" => Ok(GeneratedField::SenderPubKey),
                            "metadata" => Ok(GeneratedField::Metadata),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EncryptedPayloadEnvelope;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.EncryptedPayloadEnvelope")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<EncryptedPayloadEnvelope, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut version__ = None;
                let mut algorithm_id__ = None;
                let mut algorithm_version__ = None;
                let mut recipient_key_ids__ = None;
                let mut recipient_public_keys__ = None;
                let mut encrypted_keys__ = None;
                let mut nonce__ = None;
                let mut ciphertext__ = None;
                let mut sender_signature__ = None;
                let mut sender_pub_key__ = None;
                let mut metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AlgorithmId => {
                            if algorithm_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmId"));
                            }
                            algorithm_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AlgorithmVersion => {
                            if algorithm_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("algorithmVersion"));
                            }
                            algorithm_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RecipientKeyIds => {
                            if recipient_key_ids__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientKeyIds"));
                            }
                            recipient_key_ids__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RecipientPublicKeys => {
                            if recipient_public_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("recipientPublicKeys"));
                            }
                            recipient_public_keys__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::EncryptedKeys => {
                            if encrypted_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedKeys"));
                            }
                            encrypted_keys__ = 
                                Some(map_.next_value::<Vec<::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|x| x.0).collect())
                            ;
                        }
                        GeneratedField::Nonce => {
                            if nonce__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nonce"));
                            }
                            nonce__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Ciphertext => {
                            if ciphertext__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ciphertext"));
                            }
                            ciphertext__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SenderSignature => {
                            if sender_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("senderSignature"));
                            }
                            sender_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SenderPubKey => {
                            if sender_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("senderPubKey"));
                            }
                            sender_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(EncryptedPayloadEnvelope {
                    version: version__.unwrap_or_default(),
                    algorithm_id: algorithm_id__.unwrap_or_default(),
                    algorithm_version: algorithm_version__.unwrap_or_default(),
                    recipient_key_ids: recipient_key_ids__.unwrap_or_default(),
                    recipient_public_keys: recipient_public_keys__.unwrap_or_default(),
                    encrypted_keys: encrypted_keys__.unwrap_or_default(),
                    nonce: nonce__.unwrap_or_default(),
                    ciphertext: ciphertext__.unwrap_or_default(),
                    sender_signature: sender_signature__.unwrap_or_default(),
                    sender_pub_key: sender_pub_key__.unwrap_or_default(),
                    metadata: metadata__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.EncryptedPayloadEnvelope", FIELDS, GeneratedVisitor)
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
        if !self.identity_records.is_empty() {
            len += 1;
        }
        if !self.scopes.is_empty() {
            len += 1;
        }
        if !self.approved_clients.is_empty() {
            len += 1;
        }
        if self.params.is_some() {
            len += 1;
        }
        if !self.scores.is_empty() {
            len += 1;
        }
        if self.borderline_params.is_some() {
            len += 1;
        }
        if !self.appeal_records.is_empty() {
            len += 1;
        }
        if self.appeal_params.is_some() {
            len += 1;
        }
        if !self.compliance_records.is_empty() {
            len += 1;
        }
        if !self.compliance_providers.is_empty() {
            len += 1;
        }
        if self.compliance_params.is_some() {
            len += 1;
        }
        if !self.ml_models.is_empty() {
            len += 1;
        }
        if self.model_version_state.is_some() {
            len += 1;
        }
        if !self.model_version_history.is_empty() {
            len += 1;
        }
        if self.model_params.is_some() {
            len += 1;
        }
        if !self.pending_model_proposals.is_empty() {
            len += 1;
        }
        if !self.validator_model_reports.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.GenesisState", len)?;
        if !self.identity_records.is_empty() {
            struct_ser.serialize_field("identityRecords", &self.identity_records)?;
        }
        if !self.scopes.is_empty() {
            struct_ser.serialize_field("scopes", &self.scopes)?;
        }
        if !self.approved_clients.is_empty() {
            struct_ser.serialize_field("approvedClients", &self.approved_clients)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.scores.is_empty() {
            struct_ser.serialize_field("scores", &self.scores)?;
        }
        if let Some(v) = self.borderline_params.as_ref() {
            struct_ser.serialize_field("borderlineParams", v)?;
        }
        if !self.appeal_records.is_empty() {
            struct_ser.serialize_field("appealRecords", &self.appeal_records)?;
        }
        if let Some(v) = self.appeal_params.as_ref() {
            struct_ser.serialize_field("appealParams", v)?;
        }
        if !self.compliance_records.is_empty() {
            struct_ser.serialize_field("complianceRecords", &self.compliance_records)?;
        }
        if !self.compliance_providers.is_empty() {
            struct_ser.serialize_field("complianceProviders", &self.compliance_providers)?;
        }
        if let Some(v) = self.compliance_params.as_ref() {
            struct_ser.serialize_field("complianceParams", v)?;
        }
        if !self.ml_models.is_empty() {
            struct_ser.serialize_field("mlModels", &self.ml_models)?;
        }
        if let Some(v) = self.model_version_state.as_ref() {
            struct_ser.serialize_field("modelVersionState", v)?;
        }
        if !self.model_version_history.is_empty() {
            struct_ser.serialize_field("modelVersionHistory", &self.model_version_history)?;
        }
        if let Some(v) = self.model_params.as_ref() {
            struct_ser.serialize_field("modelParams", v)?;
        }
        if !self.pending_model_proposals.is_empty() {
            struct_ser.serialize_field("pendingModelProposals", &self.pending_model_proposals)?;
        }
        if !self.validator_model_reports.is_empty() {
            struct_ser.serialize_field("validatorModelReports", &self.validator_model_reports)?;
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
            "identity_records",
            "identityRecords",
            "scopes",
            "approved_clients",
            "approvedClients",
            "params",
            "scores",
            "borderline_params",
            "borderlineParams",
            "appeal_records",
            "appealRecords",
            "appeal_params",
            "appealParams",
            "compliance_records",
            "complianceRecords",
            "compliance_providers",
            "complianceProviders",
            "compliance_params",
            "complianceParams",
            "ml_models",
            "mlModels",
            "model_version_state",
            "modelVersionState",
            "model_version_history",
            "modelVersionHistory",
            "model_params",
            "modelParams",
            "pending_model_proposals",
            "pendingModelProposals",
            "validator_model_reports",
            "validatorModelReports",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            IdentityRecords,
            Scopes,
            ApprovedClients,
            Params,
            Scores,
            BorderlineParams,
            AppealRecords,
            AppealParams,
            ComplianceRecords,
            ComplianceProviders,
            ComplianceParams,
            MlModels,
            ModelVersionState,
            ModelVersionHistory,
            ModelParams,
            PendingModelProposals,
            ValidatorModelReports,
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
                            "identityRecords" | "identity_records" => Ok(GeneratedField::IdentityRecords),
                            "scopes" => Ok(GeneratedField::Scopes),
                            "approvedClients" | "approved_clients" => Ok(GeneratedField::ApprovedClients),
                            "params" => Ok(GeneratedField::Params),
                            "scores" => Ok(GeneratedField::Scores),
                            "borderlineParams" | "borderline_params" => Ok(GeneratedField::BorderlineParams),
                            "appealRecords" | "appeal_records" => Ok(GeneratedField::AppealRecords),
                            "appealParams" | "appeal_params" => Ok(GeneratedField::AppealParams),
                            "complianceRecords" | "compliance_records" => Ok(GeneratedField::ComplianceRecords),
                            "complianceProviders" | "compliance_providers" => Ok(GeneratedField::ComplianceProviders),
                            "complianceParams" | "compliance_params" => Ok(GeneratedField::ComplianceParams),
                            "mlModels" | "ml_models" => Ok(GeneratedField::MlModels),
                            "modelVersionState" | "model_version_state" => Ok(GeneratedField::ModelVersionState),
                            "modelVersionHistory" | "model_version_history" => Ok(GeneratedField::ModelVersionHistory),
                            "modelParams" | "model_params" => Ok(GeneratedField::ModelParams),
                            "pendingModelProposals" | "pending_model_proposals" => Ok(GeneratedField::PendingModelProposals),
                            "validatorModelReports" | "validator_model_reports" => Ok(GeneratedField::ValidatorModelReports),
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
                formatter.write_str("struct virtengine.veid.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut identity_records__ = None;
                let mut scopes__ = None;
                let mut approved_clients__ = None;
                let mut params__ = None;
                let mut scores__ = None;
                let mut borderline_params__ = None;
                let mut appeal_records__ = None;
                let mut appeal_params__ = None;
                let mut compliance_records__ = None;
                let mut compliance_providers__ = None;
                let mut compliance_params__ = None;
                let mut ml_models__ = None;
                let mut model_version_state__ = None;
                let mut model_version_history__ = None;
                let mut model_params__ = None;
                let mut pending_model_proposals__ = None;
                let mut validator_model_reports__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::IdentityRecords => {
                            if identity_records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identityRecords"));
                            }
                            identity_records__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Scopes => {
                            if scopes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopes"));
                            }
                            scopes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ApprovedClients => {
                            if approved_clients__.is_some() {
                                return Err(serde::de::Error::duplicate_field("approvedClients"));
                            }
                            approved_clients__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Scores => {
                            if scores__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scores"));
                            }
                            scores__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BorderlineParams => {
                            if borderline_params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("borderlineParams"));
                            }
                            borderline_params__ = map_.next_value()?;
                        }
                        GeneratedField::AppealRecords => {
                            if appeal_records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealRecords"));
                            }
                            appeal_records__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AppealParams => {
                            if appeal_params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealParams"));
                            }
                            appeal_params__ = map_.next_value()?;
                        }
                        GeneratedField::ComplianceRecords => {
                            if compliance_records__.is_some() {
                                return Err(serde::de::Error::duplicate_field("complianceRecords"));
                            }
                            compliance_records__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ComplianceProviders => {
                            if compliance_providers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("complianceProviders"));
                            }
                            compliance_providers__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ComplianceParams => {
                            if compliance_params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("complianceParams"));
                            }
                            compliance_params__ = map_.next_value()?;
                        }
                        GeneratedField::MlModels => {
                            if ml_models__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mlModels"));
                            }
                            ml_models__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelVersionState => {
                            if model_version_state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersionState"));
                            }
                            model_version_state__ = map_.next_value()?;
                        }
                        GeneratedField::ModelVersionHistory => {
                            if model_version_history__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersionHistory"));
                            }
                            model_version_history__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelParams => {
                            if model_params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelParams"));
                            }
                            model_params__ = map_.next_value()?;
                        }
                        GeneratedField::PendingModelProposals => {
                            if pending_model_proposals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pendingModelProposals"));
                            }
                            pending_model_proposals__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorModelReports => {
                            if validator_model_reports__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorModelReports"));
                            }
                            validator_model_reports__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GenesisState {
                    identity_records: identity_records__.unwrap_or_default(),
                    scopes: scopes__.unwrap_or_default(),
                    approved_clients: approved_clients__.unwrap_or_default(),
                    params: params__,
                    scores: scores__.unwrap_or_default(),
                    borderline_params: borderline_params__,
                    appeal_records: appeal_records__.unwrap_or_default(),
                    appeal_params: appeal_params__,
                    compliance_records: compliance_records__.unwrap_or_default(),
                    compliance_providers: compliance_providers__.unwrap_or_default(),
                    compliance_params: compliance_params__,
                    ml_models: ml_models__.unwrap_or_default(),
                    model_version_state: model_version_state__,
                    model_version_history: model_version_history__.unwrap_or_default(),
                    model_params: model_params__,
                    pending_model_proposals: pending_model_proposals__.unwrap_or_default(),
                    validator_model_reports: validator_model_reports__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for GlobalConsentUpdate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.share_with_providers {
            len += 1;
        }
        if self.share_for_verification {
            len += 1;
        }
        if self.allow_re_verification {
            len += 1;
        }
        if self.allow_derived_feature_sharing {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.GlobalConsentUpdate", len)?;
        if self.share_with_providers {
            struct_ser.serialize_field("shareWithProviders", &self.share_with_providers)?;
        }
        if self.share_for_verification {
            struct_ser.serialize_field("shareForVerification", &self.share_for_verification)?;
        }
        if self.allow_re_verification {
            struct_ser.serialize_field("allowReVerification", &self.allow_re_verification)?;
        }
        if self.allow_derived_feature_sharing {
            struct_ser.serialize_field("allowDerivedFeatureSharing", &self.allow_derived_feature_sharing)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for GlobalConsentUpdate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "share_with_providers",
            "shareWithProviders",
            "share_for_verification",
            "shareForVerification",
            "allow_re_verification",
            "allowReVerification",
            "allow_derived_feature_sharing",
            "allowDerivedFeatureSharing",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ShareWithProviders,
            ShareForVerification,
            AllowReVerification,
            AllowDerivedFeatureSharing,
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
                            "shareWithProviders" | "share_with_providers" => Ok(GeneratedField::ShareWithProviders),
                            "shareForVerification" | "share_for_verification" => Ok(GeneratedField::ShareForVerification),
                            "allowReVerification" | "allow_re_verification" => Ok(GeneratedField::AllowReVerification),
                            "allowDerivedFeatureSharing" | "allow_derived_feature_sharing" => Ok(GeneratedField::AllowDerivedFeatureSharing),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = GlobalConsentUpdate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.GlobalConsentUpdate")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GlobalConsentUpdate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut share_with_providers__ = None;
                let mut share_for_verification__ = None;
                let mut allow_re_verification__ = None;
                let mut allow_derived_feature_sharing__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ShareWithProviders => {
                            if share_with_providers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("shareWithProviders"));
                            }
                            share_with_providers__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ShareForVerification => {
                            if share_for_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("shareForVerification"));
                            }
                            share_for_verification__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowReVerification => {
                            if allow_re_verification__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowReVerification"));
                            }
                            allow_re_verification__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AllowDerivedFeatureSharing => {
                            if allow_derived_feature_sharing__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowDerivedFeatureSharing"));
                            }
                            allow_derived_feature_sharing__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(GlobalConsentUpdate {
                    share_with_providers: share_with_providers__.unwrap_or_default(),
                    share_for_verification: share_for_verification__.unwrap_or_default(),
                    allow_re_verification: allow_re_verification__.unwrap_or_default(),
                    allow_derived_feature_sharing: allow_derived_feature_sharing__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.GlobalConsentUpdate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for IdentityRecord {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.scope_refs.is_empty() {
            len += 1;
        }
        if self.current_score != 0 {
            len += 1;
        }
        if !self.score_version.is_empty() {
            len += 1;
        }
        if self.last_verified_at != 0 {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        if self.tier != 0 {
            len += 1;
        }
        if !self.flags.is_empty() {
            len += 1;
        }
        if self.locked {
            len += 1;
        }
        if !self.locked_reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.IdentityRecord", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.scope_refs.is_empty() {
            struct_ser.serialize_field("scopeRefs", &self.scope_refs)?;
        }
        if self.current_score != 0 {
            struct_ser.serialize_field("currentScore", &self.current_score)?;
        }
        if !self.score_version.is_empty() {
            struct_ser.serialize_field("scoreVersion", &self.score_version)?;
        }
        if self.last_verified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastVerifiedAt", ToString::to_string(&self.last_verified_at).as_str())?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        if self.tier != 0 {
            let v = IdentityTier::try_from(self.tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tier)))?;
            struct_ser.serialize_field("tier", &v)?;
        }
        if !self.flags.is_empty() {
            struct_ser.serialize_field("flags", &self.flags)?;
        }
        if self.locked {
            struct_ser.serialize_field("locked", &self.locked)?;
        }
        if !self.locked_reason.is_empty() {
            struct_ser.serialize_field("lockedReason", &self.locked_reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for IdentityRecord {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "scope_refs",
            "scopeRefs",
            "current_score",
            "currentScore",
            "score_version",
            "scoreVersion",
            "last_verified_at",
            "lastVerifiedAt",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "tier",
            "flags",
            "locked",
            "locked_reason",
            "lockedReason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ScopeRefs,
            CurrentScore,
            ScoreVersion,
            LastVerifiedAt,
            CreatedAt,
            UpdatedAt,
            Tier,
            Flags,
            Locked,
            LockedReason,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeRefs" | "scope_refs" => Ok(GeneratedField::ScopeRefs),
                            "currentScore" | "current_score" => Ok(GeneratedField::CurrentScore),
                            "scoreVersion" | "score_version" => Ok(GeneratedField::ScoreVersion),
                            "lastVerifiedAt" | "last_verified_at" => Ok(GeneratedField::LastVerifiedAt),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "tier" => Ok(GeneratedField::Tier),
                            "flags" => Ok(GeneratedField::Flags),
                            "locked" => Ok(GeneratedField::Locked),
                            "lockedReason" | "locked_reason" => Ok(GeneratedField::LockedReason),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IdentityRecord;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.IdentityRecord")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<IdentityRecord, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut scope_refs__ = None;
                let mut current_score__ = None;
                let mut score_version__ = None;
                let mut last_verified_at__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut tier__ = None;
                let mut flags__ = None;
                let mut locked__ = None;
                let mut locked_reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeRefs => {
                            if scope_refs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeRefs"));
                            }
                            scope_refs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CurrentScore => {
                            if current_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentScore"));
                            }
                            current_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScoreVersion => {
                            if score_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreVersion"));
                            }
                            score_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastVerifiedAt => {
                            if last_verified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastVerifiedAt"));
                            }
                            last_verified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Tier => {
                            if tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tier"));
                            }
                            tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::Flags => {
                            if flags__.is_some() {
                                return Err(serde::de::Error::duplicate_field("flags"));
                            }
                            flags__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Locked => {
                            if locked__.is_some() {
                                return Err(serde::de::Error::duplicate_field("locked"));
                            }
                            locked__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LockedReason => {
                            if locked_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lockedReason"));
                            }
                            locked_reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(IdentityRecord {
                    account_address: account_address__.unwrap_or_default(),
                    scope_refs: scope_refs__.unwrap_or_default(),
                    current_score: current_score__.unwrap_or_default(),
                    score_version: score_version__.unwrap_or_default(),
                    last_verified_at: last_verified_at__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                    tier: tier__.unwrap_or_default(),
                    flags: flags__.unwrap_or_default(),
                    locked: locked__.unwrap_or_default(),
                    locked_reason: locked_reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.IdentityRecord", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for IdentityScope {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if self.version != 0 {
            len += 1;
        }
        if self.encrypted_payload.is_some() {
            len += 1;
        }
        if self.upload_metadata.is_some() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.uploaded_at != 0 {
            len += 1;
        }
        if self.verified_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if !self.owner_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.IdentityScope", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if self.version != 0 {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if let Some(v) = self.encrypted_payload.as_ref() {
            struct_ser.serialize_field("encryptedPayload", v)?;
        }
        if let Some(v) = self.upload_metadata.as_ref() {
            struct_ser.serialize_field("uploadMetadata", v)?;
        }
        if self.status != 0 {
            let v = VerificationStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.uploaded_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("uploadedAt", ToString::to_string(&self.uploaded_at).as_str())?;
        }
        if self.verified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("verifiedAt", ToString::to_string(&self.verified_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if !self.owner_address.is_empty() {
            struct_ser.serialize_field("ownerAddress", &self.owner_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for IdentityScope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "scope_type",
            "scopeType",
            "version",
            "encrypted_payload",
            "encryptedPayload",
            "upload_metadata",
            "uploadMetadata",
            "status",
            "uploaded_at",
            "uploadedAt",
            "verified_at",
            "verifiedAt",
            "expires_at",
            "expiresAt",
            "owner_address",
            "ownerAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            ScopeType,
            Version,
            EncryptedPayload,
            UploadMetadata,
            Status,
            UploadedAt,
            VerifiedAt,
            ExpiresAt,
            OwnerAddress,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "version" => Ok(GeneratedField::Version),
                            "encryptedPayload" | "encrypted_payload" => Ok(GeneratedField::EncryptedPayload),
                            "uploadMetadata" | "upload_metadata" => Ok(GeneratedField::UploadMetadata),
                            "status" => Ok(GeneratedField::Status),
                            "uploadedAt" | "uploaded_at" => Ok(GeneratedField::UploadedAt),
                            "verifiedAt" | "verified_at" => Ok(GeneratedField::VerifiedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "ownerAddress" | "owner_address" => Ok(GeneratedField::OwnerAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IdentityScope;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.IdentityScope")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<IdentityScope, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut scope_type__ = None;
                let mut version__ = None;
                let mut encrypted_payload__ = None;
                let mut upload_metadata__ = None;
                let mut status__ = None;
                let mut uploaded_at__ = None;
                let mut verified_at__ = None;
                let mut expires_at__ = None;
                let mut owner_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EncryptedPayload => {
                            if encrypted_payload__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedPayload"));
                            }
                            encrypted_payload__ = map_.next_value()?;
                        }
                        GeneratedField::UploadMetadata => {
                            if upload_metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uploadMetadata"));
                            }
                            upload_metadata__ = map_.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::UploadedAt => {
                            if uploaded_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uploadedAt"));
                            }
                            uploaded_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VerifiedAt => {
                            if verified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedAt"));
                            }
                            verified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OwnerAddress => {
                            if owner_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ownerAddress"));
                            }
                            owner_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(IdentityScope {
                    scope_id: scope_id__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    version: version__.unwrap_or_default(),
                    encrypted_payload: encrypted_payload__,
                    upload_metadata: upload_metadata__,
                    status: status__.unwrap_or_default(),
                    uploaded_at: uploaded_at__.unwrap_or_default(),
                    verified_at: verified_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    owner_address: owner_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.IdentityScope", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for IdentityScore {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.score != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.tier != 0 {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        if self.last_updated_at != 0 {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.IdentityScore", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.score != 0 {
            struct_ser.serialize_field("score", &self.score)?;
        }
        if self.status != 0 {
            let v = AccountStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.tier != 0 {
            let v = IdentityTier::try_from(self.tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tier)))?;
            struct_ser.serialize_field("tier", &v)?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        if self.last_updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdatedAt", ToString::to_string(&self.last_updated_at).as_str())?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for IdentityScore {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "score",
            "status",
            "tier",
            "model_version",
            "modelVersion",
            "last_updated_at",
            "lastUpdatedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            Score,
            Status,
            Tier,
            ModelVersion,
            LastUpdatedAt,
            BlockHeight,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "score" => Ok(GeneratedField::Score),
                            "status" => Ok(GeneratedField::Status),
                            "tier" => Ok(GeneratedField::Tier),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            "lastUpdatedAt" | "last_updated_at" => Ok(GeneratedField::LastUpdatedAt),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IdentityScore;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.IdentityScore")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<IdentityScore, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut score__ = None;
                let mut status__ = None;
                let mut tier__ = None;
                let mut model_version__ = None;
                let mut last_updated_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::Tier => {
                            if tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tier"));
                            }
                            tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastUpdatedAt => {
                            if last_updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdatedAt"));
                            }
                            last_updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(IdentityScore {
                    account_address: account_address__.unwrap_or_default(),
                    score: score__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    tier: tier__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                    last_updated_at: last_updated_at__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.IdentityScore", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for IdentityTier {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unverified => "IDENTITY_TIER_UNVERIFIED",
            Self::Basic => "IDENTITY_TIER_BASIC",
            Self::Standard => "IDENTITY_TIER_STANDARD",
            Self::Premium => "IDENTITY_TIER_PREMIUM",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for IdentityTier {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "IDENTITY_TIER_UNVERIFIED",
            "IDENTITY_TIER_BASIC",
            "IDENTITY_TIER_STANDARD",
            "IDENTITY_TIER_PREMIUM",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IdentityTier;

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
                    "IDENTITY_TIER_UNVERIFIED" => Ok(IdentityTier::Unverified),
                    "IDENTITY_TIER_BASIC" => Ok(IdentityTier::Basic),
                    "IDENTITY_TIER_STANDARD" => Ok(IdentityTier::Standard),
                    "IDENTITY_TIER_PREMIUM" => Ok(IdentityTier::Premium),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for IdentityWallet {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.wallet_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.scope_refs.is_empty() {
            len += 1;
        }
        if self.derived_features.is_some() {
            len += 1;
        }
        if self.current_score != 0 {
            len += 1;
        }
        if self.score_status != 0 {
            len += 1;
        }
        if !self.verification_history.is_empty() {
            len += 1;
        }
        if self.consent_settings.is_some() {
            len += 1;
        }
        if !self.scope_consents.is_empty() {
            len += 1;
        }
        if !self.binding_signature.is_empty() {
            len += 1;
        }
        if !self.binding_pub_key.is_empty() {
            len += 1;
        }
        if self.last_binding_at != 0 {
            len += 1;
        }
        if self.tier != 0 {
            len += 1;
        }
        if !self.metadata.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.IdentityWallet", len)?;
        if !self.wallet_id.is_empty() {
            struct_ser.serialize_field("walletId", &self.wallet_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        if self.status != 0 {
            let v = WalletStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.scope_refs.is_empty() {
            struct_ser.serialize_field("scopeRefs", &self.scope_refs)?;
        }
        if let Some(v) = self.derived_features.as_ref() {
            struct_ser.serialize_field("derivedFeatures", v)?;
        }
        if self.current_score != 0 {
            struct_ser.serialize_field("currentScore", &self.current_score)?;
        }
        if self.score_status != 0 {
            let v = AccountStatus::try_from(self.score_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.score_status)))?;
            struct_ser.serialize_field("scoreStatus", &v)?;
        }
        if !self.verification_history.is_empty() {
            struct_ser.serialize_field("verificationHistory", &self.verification_history)?;
        }
        if let Some(v) = self.consent_settings.as_ref() {
            struct_ser.serialize_field("consentSettings", v)?;
        }
        if !self.scope_consents.is_empty() {
            struct_ser.serialize_field("scopeConsents", &self.scope_consents)?;
        }
        if !self.binding_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("bindingSignature", pbjson::private::base64::encode(&self.binding_signature).as_str())?;
        }
        if !self.binding_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("bindingPubKey", pbjson::private::base64::encode(&self.binding_pub_key).as_str())?;
        }
        if self.last_binding_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastBindingAt", ToString::to_string(&self.last_binding_at).as_str())?;
        }
        if self.tier != 0 {
            let v = IdentityTier::try_from(self.tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tier)))?;
            struct_ser.serialize_field("tier", &v)?;
        }
        if !self.metadata.is_empty() {
            struct_ser.serialize_field("metadata", &self.metadata)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for IdentityWallet {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "wallet_id",
            "walletId",
            "account_address",
            "accountAddress",
            "created_at",
            "createdAt",
            "updated_at",
            "updatedAt",
            "status",
            "scope_refs",
            "scopeRefs",
            "derived_features",
            "derivedFeatures",
            "current_score",
            "currentScore",
            "score_status",
            "scoreStatus",
            "verification_history",
            "verificationHistory",
            "consent_settings",
            "consentSettings",
            "scope_consents",
            "scopeConsents",
            "binding_signature",
            "bindingSignature",
            "binding_pub_key",
            "bindingPubKey",
            "last_binding_at",
            "lastBindingAt",
            "tier",
            "metadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            WalletId,
            AccountAddress,
            CreatedAt,
            UpdatedAt,
            Status,
            ScopeRefs,
            DerivedFeatures,
            CurrentScore,
            ScoreStatus,
            VerificationHistory,
            ConsentSettings,
            ScopeConsents,
            BindingSignature,
            BindingPubKey,
            LastBindingAt,
            Tier,
            Metadata,
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
                            "walletId" | "wallet_id" => Ok(GeneratedField::WalletId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "status" => Ok(GeneratedField::Status),
                            "scopeRefs" | "scope_refs" => Ok(GeneratedField::ScopeRefs),
                            "derivedFeatures" | "derived_features" => Ok(GeneratedField::DerivedFeatures),
                            "currentScore" | "current_score" => Ok(GeneratedField::CurrentScore),
                            "scoreStatus" | "score_status" => Ok(GeneratedField::ScoreStatus),
                            "verificationHistory" | "verification_history" => Ok(GeneratedField::VerificationHistory),
                            "consentSettings" | "consent_settings" => Ok(GeneratedField::ConsentSettings),
                            "scopeConsents" | "scope_consents" => Ok(GeneratedField::ScopeConsents),
                            "bindingSignature" | "binding_signature" => Ok(GeneratedField::BindingSignature),
                            "bindingPubKey" | "binding_pub_key" => Ok(GeneratedField::BindingPubKey),
                            "lastBindingAt" | "last_binding_at" => Ok(GeneratedField::LastBindingAt),
                            "tier" => Ok(GeneratedField::Tier),
                            "metadata" => Ok(GeneratedField::Metadata),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IdentityWallet;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.IdentityWallet")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<IdentityWallet, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut wallet_id__ = None;
                let mut account_address__ = None;
                let mut created_at__ = None;
                let mut updated_at__ = None;
                let mut status__ = None;
                let mut scope_refs__ = None;
                let mut derived_features__ = None;
                let mut current_score__ = None;
                let mut score_status__ = None;
                let mut verification_history__ = None;
                let mut consent_settings__ = None;
                let mut scope_consents__ = None;
                let mut binding_signature__ = None;
                let mut binding_pub_key__ = None;
                let mut last_binding_at__ = None;
                let mut tier__ = None;
                let mut metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::WalletId => {
                            if wallet_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("walletId"));
                            }
                            wallet_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<WalletStatus>()? as i32);
                        }
                        GeneratedField::ScopeRefs => {
                            if scope_refs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeRefs"));
                            }
                            scope_refs__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DerivedFeatures => {
                            if derived_features__.is_some() {
                                return Err(serde::de::Error::duplicate_field("derivedFeatures"));
                            }
                            derived_features__ = map_.next_value()?;
                        }
                        GeneratedField::CurrentScore => {
                            if current_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentScore"));
                            }
                            current_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScoreStatus => {
                            if score_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreStatus"));
                            }
                            score_status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::VerificationHistory => {
                            if verification_history__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verificationHistory"));
                            }
                            verification_history__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ConsentSettings => {
                            if consent_settings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentSettings"));
                            }
                            consent_settings__ = map_.next_value()?;
                        }
                        GeneratedField::ScopeConsents => {
                            if scope_consents__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeConsents"));
                            }
                            scope_consents__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                        GeneratedField::BindingSignature => {
                            if binding_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bindingSignature"));
                            }
                            binding_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BindingPubKey => {
                            if binding_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bindingPubKey"));
                            }
                            binding_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastBindingAt => {
                            if last_binding_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastBindingAt"));
                            }
                            last_binding_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Tier => {
                            if tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tier"));
                            }
                            tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(IdentityWallet {
                    wallet_id: wallet_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    scope_refs: scope_refs__.unwrap_or_default(),
                    derived_features: derived_features__,
                    current_score: current_score__.unwrap_or_default(),
                    score_status: score_status__.unwrap_or_default(),
                    verification_history: verification_history__.unwrap_or_default(),
                    consent_settings: consent_settings__,
                    scope_consents: scope_consents__.unwrap_or_default(),
                    binding_signature: binding_signature__.unwrap_or_default(),
                    binding_pub_key: binding_pub_key__.unwrap_or_default(),
                    last_binding_at: last_binding_at__.unwrap_or_default(),
                    tier: tier__.unwrap_or_default(),
                    metadata: metadata__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.IdentityWallet", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MlModelInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.model_id.is_empty() {
            len += 1;
        }
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.version.is_empty() {
            len += 1;
        }
        if !self.model_type.is_empty() {
            len += 1;
        }
        if !self.sha256_hash.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if self.activated_at != 0 {
            len += 1;
        }
        if self.registered_at != 0 {
            len += 1;
        }
        if !self.registered_by.is_empty() {
            len += 1;
        }
        if self.governance_id != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MLModelInfo", len)?;
        if !self.model_id.is_empty() {
            struct_ser.serialize_field("modelId", &self.model_id)?;
        }
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.version.is_empty() {
            struct_ser.serialize_field("version", &self.version)?;
        }
        if !self.model_type.is_empty() {
            struct_ser.serialize_field("modelType", &self.model_type)?;
        }
        if !self.sha256_hash.is_empty() {
            struct_ser.serialize_field("sha256Hash", &self.sha256_hash)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if self.activated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activatedAt", ToString::to_string(&self.activated_at).as_str())?;
        }
        if self.registered_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("registeredAt", ToString::to_string(&self.registered_at).as_str())?;
        }
        if !self.registered_by.is_empty() {
            struct_ser.serialize_field("registeredBy", &self.registered_by)?;
        }
        if self.governance_id != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("governanceId", ToString::to_string(&self.governance_id).as_str())?;
        }
        if self.status != 0 {
            let v = ModelStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MlModelInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "model_id",
            "modelId",
            "name",
            "version",
            "model_type",
            "modelType",
            "sha256_hash",
            "sha256Hash",
            "description",
            "activated_at",
            "activatedAt",
            "registered_at",
            "registeredAt",
            "registered_by",
            "registeredBy",
            "governance_id",
            "governanceId",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ModelId,
            Name,
            Version,
            ModelType,
            Sha256Hash,
            Description,
            ActivatedAt,
            RegisteredAt,
            RegisteredBy,
            GovernanceId,
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
                            "modelId" | "model_id" => Ok(GeneratedField::ModelId),
                            "name" => Ok(GeneratedField::Name),
                            "version" => Ok(GeneratedField::Version),
                            "modelType" | "model_type" => Ok(GeneratedField::ModelType),
                            "sha256Hash" | "sha256_hash" => Ok(GeneratedField::Sha256Hash),
                            "description" => Ok(GeneratedField::Description),
                            "activatedAt" | "activated_at" => Ok(GeneratedField::ActivatedAt),
                            "registeredAt" | "registered_at" => Ok(GeneratedField::RegisteredAt),
                            "registeredBy" | "registered_by" => Ok(GeneratedField::RegisteredBy),
                            "governanceId" | "governance_id" => Ok(GeneratedField::GovernanceId),
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
            type Value = MlModelInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MLModelInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MlModelInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut model_id__ = None;
                let mut name__ = None;
                let mut version__ = None;
                let mut model_type__ = None;
                let mut sha256_hash__ = None;
                let mut description__ = None;
                let mut activated_at__ = None;
                let mut registered_at__ = None;
                let mut registered_by__ = None;
                let mut governance_id__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ModelId => {
                            if model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelId"));
                            }
                            model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Version => {
                            if version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("version"));
                            }
                            version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelType => {
                            if model_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelType"));
                            }
                            model_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Sha256Hash => {
                            if sha256_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sha256Hash"));
                            }
                            sha256_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ActivatedAt => {
                            if activated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activatedAt"));
                            }
                            activated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RegisteredAt => {
                            if registered_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registeredAt"));
                            }
                            registered_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RegisteredBy => {
                            if registered_by__.is_some() {
                                return Err(serde::de::Error::duplicate_field("registeredBy"));
                            }
                            registered_by__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GovernanceId => {
                            if governance_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("governanceId"));
                            }
                            governance_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ModelStatus>()? as i32);
                        }
                    }
                }
                Ok(MlModelInfo {
                    model_id: model_id__.unwrap_or_default(),
                    name: name__.unwrap_or_default(),
                    version: version__.unwrap_or_default(),
                    model_type: model_type__.unwrap_or_default(),
                    sha256_hash: sha256_hash__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    activated_at: activated_at__.unwrap_or_default(),
                    registered_at: registered_at__.unwrap_or_default(),
                    registered_by: registered_by__.unwrap_or_default(),
                    governance_id: governance_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MLModelInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ModelParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.required_model_types.is_empty() {
            len += 1;
        }
        if self.activation_delay_blocks != 0 {
            len += 1;
        }
        if self.max_model_age_days != 0 {
            len += 1;
        }
        if !self.allowed_registrars.is_empty() {
            len += 1;
        }
        if self.validator_sync_grace_period != 0 {
            len += 1;
        }
        if self.model_update_quorum != 0 {
            len += 1;
        }
        if self.enable_governance_updates {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ModelParams", len)?;
        if !self.required_model_types.is_empty() {
            struct_ser.serialize_field("requiredModelTypes", &self.required_model_types)?;
        }
        if self.activation_delay_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activationDelayBlocks", ToString::to_string(&self.activation_delay_blocks).as_str())?;
        }
        if self.max_model_age_days != 0 {
            struct_ser.serialize_field("maxModelAgeDays", &self.max_model_age_days)?;
        }
        if !self.allowed_registrars.is_empty() {
            struct_ser.serialize_field("allowedRegistrars", &self.allowed_registrars)?;
        }
        if self.validator_sync_grace_period != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("validatorSyncGracePeriod", ToString::to_string(&self.validator_sync_grace_period).as_str())?;
        }
        if self.model_update_quorum != 0 {
            struct_ser.serialize_field("modelUpdateQuorum", &self.model_update_quorum)?;
        }
        if self.enable_governance_updates {
            struct_ser.serialize_field("enableGovernanceUpdates", &self.enable_governance_updates)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ModelParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "required_model_types",
            "requiredModelTypes",
            "activation_delay_blocks",
            "activationDelayBlocks",
            "max_model_age_days",
            "maxModelAgeDays",
            "allowed_registrars",
            "allowedRegistrars",
            "validator_sync_grace_period",
            "validatorSyncGracePeriod",
            "model_update_quorum",
            "modelUpdateQuorum",
            "enable_governance_updates",
            "enableGovernanceUpdates",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            RequiredModelTypes,
            ActivationDelayBlocks,
            MaxModelAgeDays,
            AllowedRegistrars,
            ValidatorSyncGracePeriod,
            ModelUpdateQuorum,
            EnableGovernanceUpdates,
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
                            "requiredModelTypes" | "required_model_types" => Ok(GeneratedField::RequiredModelTypes),
                            "activationDelayBlocks" | "activation_delay_blocks" => Ok(GeneratedField::ActivationDelayBlocks),
                            "maxModelAgeDays" | "max_model_age_days" => Ok(GeneratedField::MaxModelAgeDays),
                            "allowedRegistrars" | "allowed_registrars" => Ok(GeneratedField::AllowedRegistrars),
                            "validatorSyncGracePeriod" | "validator_sync_grace_period" => Ok(GeneratedField::ValidatorSyncGracePeriod),
                            "modelUpdateQuorum" | "model_update_quorum" => Ok(GeneratedField::ModelUpdateQuorum),
                            "enableGovernanceUpdates" | "enable_governance_updates" => Ok(GeneratedField::EnableGovernanceUpdates),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModelParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ModelParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ModelParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut required_model_types__ = None;
                let mut activation_delay_blocks__ = None;
                let mut max_model_age_days__ = None;
                let mut allowed_registrars__ = None;
                let mut validator_sync_grace_period__ = None;
                let mut model_update_quorum__ = None;
                let mut enable_governance_updates__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::RequiredModelTypes => {
                            if required_model_types__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requiredModelTypes"));
                            }
                            required_model_types__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ActivationDelayBlocks => {
                            if activation_delay_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activationDelayBlocks"));
                            }
                            activation_delay_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxModelAgeDays => {
                            if max_model_age_days__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxModelAgeDays"));
                            }
                            max_model_age_days__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AllowedRegistrars => {
                            if allowed_registrars__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowedRegistrars"));
                            }
                            allowed_registrars__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ValidatorSyncGracePeriod => {
                            if validator_sync_grace_period__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorSyncGracePeriod"));
                            }
                            validator_sync_grace_period__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelUpdateQuorum => {
                            if model_update_quorum__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelUpdateQuorum"));
                            }
                            model_update_quorum__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EnableGovernanceUpdates => {
                            if enable_governance_updates__.is_some() {
                                return Err(serde::de::Error::duplicate_field("enableGovernanceUpdates"));
                            }
                            enable_governance_updates__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ModelParams {
                    required_model_types: required_model_types__.unwrap_or_default(),
                    activation_delay_blocks: activation_delay_blocks__.unwrap_or_default(),
                    max_model_age_days: max_model_age_days__.unwrap_or_default(),
                    allowed_registrars: allowed_registrars__.unwrap_or_default(),
                    validator_sync_grace_period: validator_sync_grace_period__.unwrap_or_default(),
                    model_update_quorum: model_update_quorum__.unwrap_or_default(),
                    enable_governance_updates: enable_governance_updates__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ModelParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ModelProposalStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "MODEL_PROPOSAL_STATUS_UNSPECIFIED",
            Self::Pending => "MODEL_PROPOSAL_STATUS_PENDING",
            Self::Approved => "MODEL_PROPOSAL_STATUS_APPROVED",
            Self::Rejected => "MODEL_PROPOSAL_STATUS_REJECTED",
            Self::Activated => "MODEL_PROPOSAL_STATUS_ACTIVATED",
            Self::Expired => "MODEL_PROPOSAL_STATUS_EXPIRED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ModelProposalStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "MODEL_PROPOSAL_STATUS_UNSPECIFIED",
            "MODEL_PROPOSAL_STATUS_PENDING",
            "MODEL_PROPOSAL_STATUS_APPROVED",
            "MODEL_PROPOSAL_STATUS_REJECTED",
            "MODEL_PROPOSAL_STATUS_ACTIVATED",
            "MODEL_PROPOSAL_STATUS_EXPIRED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModelProposalStatus;

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
                    "MODEL_PROPOSAL_STATUS_UNSPECIFIED" => Ok(ModelProposalStatus::Unspecified),
                    "MODEL_PROPOSAL_STATUS_PENDING" => Ok(ModelProposalStatus::Pending),
                    "MODEL_PROPOSAL_STATUS_APPROVED" => Ok(ModelProposalStatus::Approved),
                    "MODEL_PROPOSAL_STATUS_REJECTED" => Ok(ModelProposalStatus::Rejected),
                    "MODEL_PROPOSAL_STATUS_ACTIVATED" => Ok(ModelProposalStatus::Activated),
                    "MODEL_PROPOSAL_STATUS_EXPIRED" => Ok(ModelProposalStatus::Expired),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ModelStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "MODEL_STATUS_UNSPECIFIED",
            Self::Pending => "MODEL_STATUS_PENDING",
            Self::Active => "MODEL_STATUS_ACTIVE",
            Self::Deprecated => "MODEL_STATUS_DEPRECATED",
            Self::Revoked => "MODEL_STATUS_REVOKED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ModelStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "MODEL_STATUS_UNSPECIFIED",
            "MODEL_STATUS_PENDING",
            "MODEL_STATUS_ACTIVE",
            "MODEL_STATUS_DEPRECATED",
            "MODEL_STATUS_REVOKED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModelStatus;

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
                    "MODEL_STATUS_UNSPECIFIED" => Ok(ModelStatus::Unspecified),
                    "MODEL_STATUS_PENDING" => Ok(ModelStatus::Pending),
                    "MODEL_STATUS_ACTIVE" => Ok(ModelStatus::Active),
                    "MODEL_STATUS_DEPRECATED" => Ok(ModelStatus::Deprecated),
                    "MODEL_STATUS_REVOKED" => Ok(ModelStatus::Revoked),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ModelType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "MODEL_TYPE_UNSPECIFIED",
            Self::TrustScore => "MODEL_TYPE_TRUST_SCORE",
            Self::FaceVerification => "MODEL_TYPE_FACE_VERIFICATION",
            Self::Liveness => "MODEL_TYPE_LIVENESS",
            Self::GanDetection => "MODEL_TYPE_GAN_DETECTION",
            Self::Ocr => "MODEL_TYPE_OCR",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ModelType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "MODEL_TYPE_UNSPECIFIED",
            "MODEL_TYPE_TRUST_SCORE",
            "MODEL_TYPE_FACE_VERIFICATION",
            "MODEL_TYPE_LIVENESS",
            "MODEL_TYPE_GAN_DETECTION",
            "MODEL_TYPE_OCR",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModelType;

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
                    "MODEL_TYPE_UNSPECIFIED" => Ok(ModelType::Unspecified),
                    "MODEL_TYPE_TRUST_SCORE" => Ok(ModelType::TrustScore),
                    "MODEL_TYPE_FACE_VERIFICATION" => Ok(ModelType::FaceVerification),
                    "MODEL_TYPE_LIVENESS" => Ok(ModelType::Liveness),
                    "MODEL_TYPE_GAN_DETECTION" => Ok(ModelType::GanDetection),
                    "MODEL_TYPE_OCR" => Ok(ModelType::Ocr),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ModelUpdateProposal {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.title.is_empty() {
            len += 1;
        }
        if !self.description.is_empty() {
            len += 1;
        }
        if !self.model_type.is_empty() {
            len += 1;
        }
        if !self.new_model_id.is_empty() {
            len += 1;
        }
        if !self.new_model_hash.is_empty() {
            len += 1;
        }
        if self.activation_delay != 0 {
            len += 1;
        }
        if self.proposed_at != 0 {
            len += 1;
        }
        if !self.proposer_address.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.governance_id != 0 {
            len += 1;
        }
        if self.activation_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ModelUpdateProposal", len)?;
        if !self.title.is_empty() {
            struct_ser.serialize_field("title", &self.title)?;
        }
        if !self.description.is_empty() {
            struct_ser.serialize_field("description", &self.description)?;
        }
        if !self.model_type.is_empty() {
            struct_ser.serialize_field("modelType", &self.model_type)?;
        }
        if !self.new_model_id.is_empty() {
            struct_ser.serialize_field("newModelId", &self.new_model_id)?;
        }
        if !self.new_model_hash.is_empty() {
            struct_ser.serialize_field("newModelHash", &self.new_model_hash)?;
        }
        if self.activation_delay != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activationDelay", ToString::to_string(&self.activation_delay).as_str())?;
        }
        if self.proposed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("proposedAt", ToString::to_string(&self.proposed_at).as_str())?;
        }
        if !self.proposer_address.is_empty() {
            struct_ser.serialize_field("proposerAddress", &self.proposer_address)?;
        }
        if self.status != 0 {
            let v = ModelProposalStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.governance_id != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("governanceId", ToString::to_string(&self.governance_id).as_str())?;
        }
        if self.activation_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activationHeight", ToString::to_string(&self.activation_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ModelUpdateProposal {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "title",
            "description",
            "model_type",
            "modelType",
            "new_model_id",
            "newModelId",
            "new_model_hash",
            "newModelHash",
            "activation_delay",
            "activationDelay",
            "proposed_at",
            "proposedAt",
            "proposer_address",
            "proposerAddress",
            "status",
            "governance_id",
            "governanceId",
            "activation_height",
            "activationHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Title,
            Description,
            ModelType,
            NewModelId,
            NewModelHash,
            ActivationDelay,
            ProposedAt,
            ProposerAddress,
            Status,
            GovernanceId,
            ActivationHeight,
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
                            "title" => Ok(GeneratedField::Title),
                            "description" => Ok(GeneratedField::Description),
                            "modelType" | "model_type" => Ok(GeneratedField::ModelType),
                            "newModelId" | "new_model_id" => Ok(GeneratedField::NewModelId),
                            "newModelHash" | "new_model_hash" => Ok(GeneratedField::NewModelHash),
                            "activationDelay" | "activation_delay" => Ok(GeneratedField::ActivationDelay),
                            "proposedAt" | "proposed_at" => Ok(GeneratedField::ProposedAt),
                            "proposerAddress" | "proposer_address" => Ok(GeneratedField::ProposerAddress),
                            "status" => Ok(GeneratedField::Status),
                            "governanceId" | "governance_id" => Ok(GeneratedField::GovernanceId),
                            "activationHeight" | "activation_height" => Ok(GeneratedField::ActivationHeight),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModelUpdateProposal;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ModelUpdateProposal")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ModelUpdateProposal, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut title__ = None;
                let mut description__ = None;
                let mut model_type__ = None;
                let mut new_model_id__ = None;
                let mut new_model_hash__ = None;
                let mut activation_delay__ = None;
                let mut proposed_at__ = None;
                let mut proposer_address__ = None;
                let mut status__ = None;
                let mut governance_id__ = None;
                let mut activation_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Title => {
                            if title__.is_some() {
                                return Err(serde::de::Error::duplicate_field("title"));
                            }
                            title__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Description => {
                            if description__.is_some() {
                                return Err(serde::de::Error::duplicate_field("description"));
                            }
                            description__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelType => {
                            if model_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelType"));
                            }
                            model_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewModelId => {
                            if new_model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newModelId"));
                            }
                            new_model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewModelHash => {
                            if new_model_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newModelHash"));
                            }
                            new_model_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ActivationDelay => {
                            if activation_delay__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activationDelay"));
                            }
                            activation_delay__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ProposedAt => {
                            if proposed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposedAt"));
                            }
                            proposed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ProposerAddress => {
                            if proposer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposerAddress"));
                            }
                            proposer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ModelProposalStatus>()? as i32);
                        }
                        GeneratedField::GovernanceId => {
                            if governance_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("governanceId"));
                            }
                            governance_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ActivationHeight => {
                            if activation_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activationHeight"));
                            }
                            activation_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ModelUpdateProposal {
                    title: title__.unwrap_or_default(),
                    description: description__.unwrap_or_default(),
                    model_type: model_type__.unwrap_or_default(),
                    new_model_id: new_model_id__.unwrap_or_default(),
                    new_model_hash: new_model_hash__.unwrap_or_default(),
                    activation_delay: activation_delay__.unwrap_or_default(),
                    proposed_at: proposed_at__.unwrap_or_default(),
                    proposer_address: proposer_address__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    governance_id: governance_id__.unwrap_or_default(),
                    activation_height: activation_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ModelUpdateProposal", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ModelVersionHistory {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.history_id.is_empty() {
            len += 1;
        }
        if !self.model_type.is_empty() {
            len += 1;
        }
        if !self.old_model_id.is_empty() {
            len += 1;
        }
        if !self.new_model_id.is_empty() {
            len += 1;
        }
        if !self.old_model_hash.is_empty() {
            len += 1;
        }
        if !self.new_model_hash.is_empty() {
            len += 1;
        }
        if self.changed_at != 0 {
            len += 1;
        }
        if self.governance_id != 0 {
            len += 1;
        }
        if !self.proposer_address.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ModelVersionHistory", len)?;
        if !self.history_id.is_empty() {
            struct_ser.serialize_field("historyId", &self.history_id)?;
        }
        if !self.model_type.is_empty() {
            struct_ser.serialize_field("modelType", &self.model_type)?;
        }
        if !self.old_model_id.is_empty() {
            struct_ser.serialize_field("oldModelId", &self.old_model_id)?;
        }
        if !self.new_model_id.is_empty() {
            struct_ser.serialize_field("newModelId", &self.new_model_id)?;
        }
        if !self.old_model_hash.is_empty() {
            struct_ser.serialize_field("oldModelHash", &self.old_model_hash)?;
        }
        if !self.new_model_hash.is_empty() {
            struct_ser.serialize_field("newModelHash", &self.new_model_hash)?;
        }
        if self.changed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("changedAt", ToString::to_string(&self.changed_at).as_str())?;
        }
        if self.governance_id != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("governanceId", ToString::to_string(&self.governance_id).as_str())?;
        }
        if !self.proposer_address.is_empty() {
            struct_ser.serialize_field("proposerAddress", &self.proposer_address)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ModelVersionHistory {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "history_id",
            "historyId",
            "model_type",
            "modelType",
            "old_model_id",
            "oldModelId",
            "new_model_id",
            "newModelId",
            "old_model_hash",
            "oldModelHash",
            "new_model_hash",
            "newModelHash",
            "changed_at",
            "changedAt",
            "governance_id",
            "governanceId",
            "proposer_address",
            "proposerAddress",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            HistoryId,
            ModelType,
            OldModelId,
            NewModelId,
            OldModelHash,
            NewModelHash,
            ChangedAt,
            GovernanceId,
            ProposerAddress,
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
                            "historyId" | "history_id" => Ok(GeneratedField::HistoryId),
                            "modelType" | "model_type" => Ok(GeneratedField::ModelType),
                            "oldModelId" | "old_model_id" => Ok(GeneratedField::OldModelId),
                            "newModelId" | "new_model_id" => Ok(GeneratedField::NewModelId),
                            "oldModelHash" | "old_model_hash" => Ok(GeneratedField::OldModelHash),
                            "newModelHash" | "new_model_hash" => Ok(GeneratedField::NewModelHash),
                            "changedAt" | "changed_at" => Ok(GeneratedField::ChangedAt),
                            "governanceId" | "governance_id" => Ok(GeneratedField::GovernanceId),
                            "proposerAddress" | "proposer_address" => Ok(GeneratedField::ProposerAddress),
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
            type Value = ModelVersionHistory;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ModelVersionHistory")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ModelVersionHistory, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut history_id__ = None;
                let mut model_type__ = None;
                let mut old_model_id__ = None;
                let mut new_model_id__ = None;
                let mut old_model_hash__ = None;
                let mut new_model_hash__ = None;
                let mut changed_at__ = None;
                let mut governance_id__ = None;
                let mut proposer_address__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::HistoryId => {
                            if history_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("historyId"));
                            }
                            history_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelType => {
                            if model_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelType"));
                            }
                            model_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OldModelId => {
                            if old_model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldModelId"));
                            }
                            old_model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewModelId => {
                            if new_model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newModelId"));
                            }
                            new_model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OldModelHash => {
                            if old_model_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldModelHash"));
                            }
                            old_model_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewModelHash => {
                            if new_model_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newModelHash"));
                            }
                            new_model_hash__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChangedAt => {
                            if changed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("changedAt"));
                            }
                            changed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GovernanceId => {
                            if governance_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("governanceId"));
                            }
                            governance_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ProposerAddress => {
                            if proposer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposerAddress"));
                            }
                            proposer_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ModelVersionHistory {
                    history_id: history_id__.unwrap_or_default(),
                    model_type: model_type__.unwrap_or_default(),
                    old_model_id: old_model_id__.unwrap_or_default(),
                    new_model_id: new_model_id__.unwrap_or_default(),
                    old_model_hash: old_model_hash__.unwrap_or_default(),
                    new_model_hash: new_model_hash__.unwrap_or_default(),
                    changed_at: changed_at__.unwrap_or_default(),
                    governance_id: governance_id__.unwrap_or_default(),
                    proposer_address: proposer_address__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ModelVersionHistory", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ModelVersionState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.trust_score_model.is_empty() {
            len += 1;
        }
        if !self.face_verification_model.is_empty() {
            len += 1;
        }
        if !self.liveness_model.is_empty() {
            len += 1;
        }
        if !self.gan_detection_model.is_empty() {
            len += 1;
        }
        if !self.ocr_model.is_empty() {
            len += 1;
        }
        if self.last_updated != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ModelVersionState", len)?;
        if !self.trust_score_model.is_empty() {
            struct_ser.serialize_field("trustScoreModel", &self.trust_score_model)?;
        }
        if !self.face_verification_model.is_empty() {
            struct_ser.serialize_field("faceVerificationModel", &self.face_verification_model)?;
        }
        if !self.liveness_model.is_empty() {
            struct_ser.serialize_field("livenessModel", &self.liveness_model)?;
        }
        if !self.gan_detection_model.is_empty() {
            struct_ser.serialize_field("ganDetectionModel", &self.gan_detection_model)?;
        }
        if !self.ocr_model.is_empty() {
            struct_ser.serialize_field("ocrModel", &self.ocr_model)?;
        }
        if self.last_updated != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdated", ToString::to_string(&self.last_updated).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ModelVersionState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "trust_score_model",
            "trustScoreModel",
            "face_verification_model",
            "faceVerificationModel",
            "liveness_model",
            "livenessModel",
            "gan_detection_model",
            "ganDetectionModel",
            "ocr_model",
            "ocrModel",
            "last_updated",
            "lastUpdated",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TrustScoreModel,
            FaceVerificationModel,
            LivenessModel,
            GanDetectionModel,
            OcrModel,
            LastUpdated,
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
                            "trustScoreModel" | "trust_score_model" => Ok(GeneratedField::TrustScoreModel),
                            "faceVerificationModel" | "face_verification_model" => Ok(GeneratedField::FaceVerificationModel),
                            "livenessModel" | "liveness_model" => Ok(GeneratedField::LivenessModel),
                            "ganDetectionModel" | "gan_detection_model" => Ok(GeneratedField::GanDetectionModel),
                            "ocrModel" | "ocr_model" => Ok(GeneratedField::OcrModel),
                            "lastUpdated" | "last_updated" => Ok(GeneratedField::LastUpdated),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ModelVersionState;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ModelVersionState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ModelVersionState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut trust_score_model__ = None;
                let mut face_verification_model__ = None;
                let mut liveness_model__ = None;
                let mut gan_detection_model__ = None;
                let mut ocr_model__ = None;
                let mut last_updated__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::TrustScoreModel => {
                            if trust_score_model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("trustScoreModel"));
                            }
                            trust_score_model__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FaceVerificationModel => {
                            if face_verification_model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("faceVerificationModel"));
                            }
                            face_verification_model__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LivenessModel => {
                            if liveness_model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("livenessModel"));
                            }
                            liveness_model__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GanDetectionModel => {
                            if gan_detection_model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ganDetectionModel"));
                            }
                            gan_detection_model__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OcrModel => {
                            if ocr_model__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ocrModel"));
                            }
                            ocr_model__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastUpdated => {
                            if last_updated__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdated"));
                            }
                            last_updated__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ModelVersionState {
                    trust_score_model: trust_score_model__.unwrap_or_default(),
                    face_verification_model: face_verification_model__.unwrap_or_default(),
                    liveness_model: liveness_model__.unwrap_or_default(),
                    gan_detection_model: gan_detection_model__.unwrap_or_default(),
                    ocr_model: ocr_model__.unwrap_or_default(),
                    last_updated: last_updated__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ModelVersionState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgActivateModel {
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
        if !self.model_type.is_empty() {
            len += 1;
        }
        if !self.model_id.is_empty() {
            len += 1;
        }
        if self.governance_id != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgActivateModel", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.model_type.is_empty() {
            struct_ser.serialize_field("modelType", &self.model_type)?;
        }
        if !self.model_id.is_empty() {
            struct_ser.serialize_field("modelId", &self.model_id)?;
        }
        if self.governance_id != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("governanceId", ToString::to_string(&self.governance_id).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgActivateModel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "model_type",
            "modelType",
            "model_id",
            "modelId",
            "governance_id",
            "governanceId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ModelType,
            ModelId,
            GovernanceId,
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
                            "modelType" | "model_type" => Ok(GeneratedField::ModelType),
                            "modelId" | "model_id" => Ok(GeneratedField::ModelId),
                            "governanceId" | "governance_id" => Ok(GeneratedField::GovernanceId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgActivateModel;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgActivateModel")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgActivateModel, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut model_type__ = None;
                let mut model_id__ = None;
                let mut governance_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelType => {
                            if model_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelType"));
                            }
                            model_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelId => {
                            if model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelId"));
                            }
                            model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GovernanceId => {
                            if governance_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("governanceId"));
                            }
                            governance_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgActivateModel {
                    authority: authority__.unwrap_or_default(),
                    model_type: model_type__.unwrap_or_default(),
                    model_id: model_id__.unwrap_or_default(),
                    governance_id: governance_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgActivateModel", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgActivateModelResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.activated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgActivateModelResponse", len)?;
        if self.activated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("activatedAt", ToString::to_string(&self.activated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgActivateModelResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "activated_at",
            "activatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ActivatedAt,
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
                            "activatedAt" | "activated_at" => Ok(GeneratedField::ActivatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgActivateModelResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgActivateModelResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgActivateModelResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut activated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ActivatedAt => {
                            if activated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activatedAt"));
                            }
                            activated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgActivateModelResponse {
                    activated_at: activated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgActivateModelResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAddScopeToWallet {
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
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if !self.envelope_hash.is_empty() {
            len += 1;
        }
        if !self.user_signature.is_empty() {
            len += 1;
        }
        if self.grant_consent {
            len += 1;
        }
        if !self.consent_purpose.is_empty() {
            len += 1;
        }
        if self.consent_expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgAddScopeToWallet", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if !self.envelope_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("envelopeHash", pbjson::private::base64::encode(&self.envelope_hash).as_str())?;
        }
        if !self.user_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userSignature", pbjson::private::base64::encode(&self.user_signature).as_str())?;
        }
        if self.grant_consent {
            struct_ser.serialize_field("grantConsent", &self.grant_consent)?;
        }
        if !self.consent_purpose.is_empty() {
            struct_ser.serialize_field("consentPurpose", &self.consent_purpose)?;
        }
        if self.consent_expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("consentExpiresAt", ToString::to_string(&self.consent_expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAddScopeToWallet {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "scope_id",
            "scopeId",
            "scope_type",
            "scopeType",
            "envelope_hash",
            "envelopeHash",
            "user_signature",
            "userSignature",
            "grant_consent",
            "grantConsent",
            "consent_purpose",
            "consentPurpose",
            "consent_expires_at",
            "consentExpiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ScopeId,
            ScopeType,
            EnvelopeHash,
            UserSignature,
            GrantConsent,
            ConsentPurpose,
            ConsentExpiresAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "envelopeHash" | "envelope_hash" => Ok(GeneratedField::EnvelopeHash),
                            "userSignature" | "user_signature" => Ok(GeneratedField::UserSignature),
                            "grantConsent" | "grant_consent" => Ok(GeneratedField::GrantConsent),
                            "consentPurpose" | "consent_purpose" => Ok(GeneratedField::ConsentPurpose),
                            "consentExpiresAt" | "consent_expires_at" => Ok(GeneratedField::ConsentExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAddScopeToWallet;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgAddScopeToWallet")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAddScopeToWallet, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut scope_id__ = None;
                let mut scope_type__ = None;
                let mut envelope_hash__ = None;
                let mut user_signature__ = None;
                let mut grant_consent__ = None;
                let mut consent_purpose__ = None;
                let mut consent_expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::EnvelopeHash => {
                            if envelope_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("envelopeHash"));
                            }
                            envelope_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UserSignature => {
                            if user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userSignature"));
                            }
                            user_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GrantConsent => {
                            if grant_consent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("grantConsent"));
                            }
                            grant_consent__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ConsentPurpose => {
                            if consent_purpose__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentPurpose"));
                            }
                            consent_purpose__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ConsentExpiresAt => {
                            if consent_expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentExpiresAt"));
                            }
                            consent_expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAddScopeToWallet {
                    sender: sender__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    envelope_hash: envelope_hash__.unwrap_or_default(),
                    user_signature: user_signature__.unwrap_or_default(),
                    grant_consent: grant_consent__.unwrap_or_default(),
                    consent_purpose: consent_purpose__.unwrap_or_default(),
                    consent_expires_at: consent_expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgAddScopeToWallet", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAddScopeToWalletResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.added_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgAddScopeToWalletResponse", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.added_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("addedAt", ToString::to_string(&self.added_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAddScopeToWalletResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "added_at",
            "addedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            AddedAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "addedAt" | "added_at" => Ok(GeneratedField::AddedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAddScopeToWalletResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgAddScopeToWalletResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAddScopeToWalletResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut added_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AddedAt => {
                            if added_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addedAt"));
                            }
                            added_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAddScopeToWalletResponse {
                    scope_id: scope_id__.unwrap_or_default(),
                    added_at: added_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgAddScopeToWalletResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAttestCompliance {
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
        if !self.target_address.is_empty() {
            len += 1;
        }
        if !self.attestation_type.is_empty() {
            len += 1;
        }
        if self.expiry_blocks != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgAttestCompliance", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.target_address.is_empty() {
            struct_ser.serialize_field("targetAddress", &self.target_address)?;
        }
        if !self.attestation_type.is_empty() {
            struct_ser.serialize_field("attestationType", &self.attestation_type)?;
        }
        if self.expiry_blocks != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiryBlocks", ToString::to_string(&self.expiry_blocks).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAttestCompliance {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "target_address",
            "targetAddress",
            "attestation_type",
            "attestationType",
            "expiry_blocks",
            "expiryBlocks",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            TargetAddress,
            AttestationType,
            ExpiryBlocks,
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
                            "targetAddress" | "target_address" => Ok(GeneratedField::TargetAddress),
                            "attestationType" | "attestation_type" => Ok(GeneratedField::AttestationType),
                            "expiryBlocks" | "expiry_blocks" => Ok(GeneratedField::ExpiryBlocks),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAttestCompliance;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgAttestCompliance")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAttestCompliance, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut target_address__ = None;
                let mut attestation_type__ = None;
                let mut expiry_blocks__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TargetAddress => {
                            if target_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("targetAddress"));
                            }
                            target_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AttestationType => {
                            if attestation_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestationType"));
                            }
                            attestation_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiryBlocks => {
                            if expiry_blocks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiryBlocks"));
                            }
                            expiry_blocks__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAttestCompliance {
                    validator_address: validator_address__.unwrap_or_default(),
                    target_address: target_address__.unwrap_or_default(),
                    attestation_type: attestation_type__.unwrap_or_default(),
                    expiry_blocks: expiry_blocks__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgAttestCompliance", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgAttestComplianceResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.attested_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgAttestComplianceResponse", len)?;
        if self.attested_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("attestedAt", ToString::to_string(&self.attested_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgAttestComplianceResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "attested_at",
            "attestedAt",
            "expires_at",
            "expiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AttestedAt,
            ExpiresAt,
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
                            "attestedAt" | "attested_at" => Ok(GeneratedField::AttestedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgAttestComplianceResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgAttestComplianceResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgAttestComplianceResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut attested_at__ = None;
                let mut expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AttestedAt => {
                            if attested_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("attestedAt"));
                            }
                            attested_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgAttestComplianceResponse {
                    attested_at: attested_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgAttestComplianceResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimAppeal {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.reviewer.is_empty() {
            len += 1;
        }
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgClaimAppeal", len)?;
        if !self.reviewer.is_empty() {
            struct_ser.serialize_field("reviewer", &self.reviewer)?;
        }
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimAppeal {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reviewer",
            "appeal_id",
            "appealId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reviewer,
            AppealId,
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
                            "reviewer" => Ok(GeneratedField::Reviewer),
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgClaimAppeal;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgClaimAppeal")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimAppeal, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reviewer__ = None;
                let mut appeal_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reviewer => {
                            if reviewer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewer"));
                            }
                            reviewer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgClaimAppeal {
                    reviewer: reviewer__.unwrap_or_default(),
                    appeal_id: appeal_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgClaimAppeal", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgClaimAppealResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        if self.claimed_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgClaimAppealResponse", len)?;
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        if self.claimed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("claimedAt", ToString::to_string(&self.claimed_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgClaimAppealResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_id",
            "appealId",
            "claimed_at",
            "claimedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealId,
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
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
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
            type Value = MsgClaimAppealResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgClaimAppealResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgClaimAppealResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_id__ = None;
                let mut claimed_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClaimedAt => {
                            if claimed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("claimedAt"));
                            }
                            claimed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgClaimAppealResponse {
                    appeal_id: appeal_id__.unwrap_or_default(),
                    claimed_at: claimed_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgClaimAppealResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCompleteBorderlineFallback {
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
        if !self.challenge_id.is_empty() {
            len += 1;
        }
        if !self.factors_satisfied.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgCompleteBorderlineFallback", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.challenge_id.is_empty() {
            struct_ser.serialize_field("challengeId", &self.challenge_id)?;
        }
        if !self.factors_satisfied.is_empty() {
            struct_ser.serialize_field("factorsSatisfied", &self.factors_satisfied)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCompleteBorderlineFallback {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "challenge_id",
            "challengeId",
            "factors_satisfied",
            "factorsSatisfied",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ChallengeId,
            FactorsSatisfied,
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
                            "challengeId" | "challenge_id" => Ok(GeneratedField::ChallengeId),
                            "factorsSatisfied" | "factors_satisfied" => Ok(GeneratedField::FactorsSatisfied),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCompleteBorderlineFallback;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgCompleteBorderlineFallback")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCompleteBorderlineFallback, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut challenge_id__ = None;
                let mut factors_satisfied__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ChallengeId => {
                            if challenge_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("challengeId"));
                            }
                            challenge_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FactorsSatisfied => {
                            if factors_satisfied__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorsSatisfied"));
                            }
                            factors_satisfied__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgCompleteBorderlineFallback {
                    sender: sender__.unwrap_or_default(),
                    challenge_id: challenge_id__.unwrap_or_default(),
                    factors_satisfied: factors_satisfied__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgCompleteBorderlineFallback", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCompleteBorderlineFallbackResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.fallback_id.is_empty() {
            len += 1;
        }
        if self.final_status != 0 {
            len += 1;
        }
        if !self.factor_class.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgCompleteBorderlineFallbackResponse", len)?;
        if !self.fallback_id.is_empty() {
            struct_ser.serialize_field("fallbackId", &self.fallback_id)?;
        }
        if self.final_status != 0 {
            let v = VerificationStatus::try_from(self.final_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.final_status)))?;
            struct_ser.serialize_field("finalStatus", &v)?;
        }
        if !self.factor_class.is_empty() {
            struct_ser.serialize_field("factorClass", &self.factor_class)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCompleteBorderlineFallbackResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "fallback_id",
            "fallbackId",
            "final_status",
            "finalStatus",
            "factor_class",
            "factorClass",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            FallbackId,
            FinalStatus,
            FactorClass,
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
                            "fallbackId" | "fallback_id" => Ok(GeneratedField::FallbackId),
                            "finalStatus" | "final_status" => Ok(GeneratedField::FinalStatus),
                            "factorClass" | "factor_class" => Ok(GeneratedField::FactorClass),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCompleteBorderlineFallbackResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgCompleteBorderlineFallbackResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCompleteBorderlineFallbackResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut fallback_id__ = None;
                let mut final_status__ = None;
                let mut factor_class__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::FallbackId => {
                            if fallback_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("fallbackId"));
                            }
                            fallback_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FinalStatus => {
                            if final_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("finalStatus"));
                            }
                            final_status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::FactorClass => {
                            if factor_class__.is_some() {
                                return Err(serde::de::Error::duplicate_field("factorClass"));
                            }
                            factor_class__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgCompleteBorderlineFallbackResponse {
                    fallback_id: fallback_id__.unwrap_or_default(),
                    final_status: final_status__.unwrap_or_default(),
                    factor_class: factor_class__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgCompleteBorderlineFallbackResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateIdentityWallet {
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
        if !self.binding_signature.is_empty() {
            len += 1;
        }
        if !self.binding_pub_key.is_empty() {
            len += 1;
        }
        if self.initial_consent.is_some() {
            len += 1;
        }
        if !self.metadata.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgCreateIdentityWallet", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.binding_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("bindingSignature", pbjson::private::base64::encode(&self.binding_signature).as_str())?;
        }
        if !self.binding_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("bindingPubKey", pbjson::private::base64::encode(&self.binding_pub_key).as_str())?;
        }
        if let Some(v) = self.initial_consent.as_ref() {
            struct_ser.serialize_field("initialConsent", v)?;
        }
        if !self.metadata.is_empty() {
            struct_ser.serialize_field("metadata", &self.metadata)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateIdentityWallet {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "binding_signature",
            "bindingSignature",
            "binding_pub_key",
            "bindingPubKey",
            "initial_consent",
            "initialConsent",
            "metadata",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            BindingSignature,
            BindingPubKey,
            InitialConsent,
            Metadata,
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
                            "bindingSignature" | "binding_signature" => Ok(GeneratedField::BindingSignature),
                            "bindingPubKey" | "binding_pub_key" => Ok(GeneratedField::BindingPubKey),
                            "initialConsent" | "initial_consent" => Ok(GeneratedField::InitialConsent),
                            "metadata" => Ok(GeneratedField::Metadata),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateIdentityWallet;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgCreateIdentityWallet")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateIdentityWallet, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut binding_signature__ = None;
                let mut binding_pub_key__ = None;
                let mut initial_consent__ = None;
                let mut metadata__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::BindingSignature => {
                            if binding_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bindingSignature"));
                            }
                            binding_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BindingPubKey => {
                            if binding_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bindingPubKey"));
                            }
                            binding_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::InitialConsent => {
                            if initial_consent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("initialConsent"));
                            }
                            initial_consent__ = map_.next_value()?;
                        }
                        GeneratedField::Metadata => {
                            if metadata__.is_some() {
                                return Err(serde::de::Error::duplicate_field("metadata"));
                            }
                            metadata__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(MsgCreateIdentityWallet {
                    sender: sender__.unwrap_or_default(),
                    binding_signature: binding_signature__.unwrap_or_default(),
                    binding_pub_key: binding_pub_key__.unwrap_or_default(),
                    initial_consent: initial_consent__,
                    metadata: metadata__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgCreateIdentityWallet", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgCreateIdentityWalletResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.wallet_id.is_empty() {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgCreateIdentityWalletResponse", len)?;
        if !self.wallet_id.is_empty() {
            struct_ser.serialize_field("walletId", &self.wallet_id)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgCreateIdentityWalletResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "wallet_id",
            "walletId",
            "created_at",
            "createdAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            WalletId,
            CreatedAt,
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
                            "walletId" | "wallet_id" => Ok(GeneratedField::WalletId),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgCreateIdentityWalletResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgCreateIdentityWalletResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgCreateIdentityWalletResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut wallet_id__ = None;
                let mut created_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::WalletId => {
                            if wallet_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("walletId"));
                            }
                            wallet_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgCreateIdentityWalletResponse {
                    wallet_id: wallet_id__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgCreateIdentityWalletResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeactivateComplianceProvider {
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
        if !self.provider_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgDeactivateComplianceProvider", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.provider_id.is_empty() {
            struct_ser.serialize_field("providerId", &self.provider_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeactivateComplianceProvider {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "provider_id",
            "providerId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ProviderId,
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
                            "authority" => Ok(GeneratedField::Authority),
                            "providerId" | "provider_id" => Ok(GeneratedField::ProviderId),
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
            type Value = MsgDeactivateComplianceProvider;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgDeactivateComplianceProvider")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeactivateComplianceProvider, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut provider_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderId => {
                            if provider_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerId"));
                            }
                            provider_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDeactivateComplianceProvider {
                    authority: authority__.unwrap_or_default(),
                    provider_id: provider_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgDeactivateComplianceProvider", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeactivateComplianceProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgDeactivateComplianceProviderResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeactivateComplianceProviderResponse {
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
            type Value = MsgDeactivateComplianceProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgDeactivateComplianceProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeactivateComplianceProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDeactivateComplianceProviderResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgDeactivateComplianceProviderResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeprecateModel {
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
        if !self.model_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgDeprecateModel", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.model_id.is_empty() {
            struct_ser.serialize_field("modelId", &self.model_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeprecateModel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "model_id",
            "modelId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ModelId,
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
                            "authority" => Ok(GeneratedField::Authority),
                            "modelId" | "model_id" => Ok(GeneratedField::ModelId),
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
            type Value = MsgDeprecateModel;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgDeprecateModel")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeprecateModel, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut model_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelId => {
                            if model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelId"));
                            }
                            model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDeprecateModel {
                    authority: authority__.unwrap_or_default(),
                    model_id: model_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgDeprecateModel", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeprecateModelResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgDeprecateModelResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeprecateModelResponse {
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
            type Value = MsgDeprecateModelResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgDeprecateModelResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeprecateModelResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDeprecateModelResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgDeprecateModelResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgProposeModelUpdate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.proposer.is_empty() {
            len += 1;
        }
        if self.proposal.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgProposeModelUpdate", len)?;
        if !self.proposer.is_empty() {
            struct_ser.serialize_field("proposer", &self.proposer)?;
        }
        if let Some(v) = self.proposal.as_ref() {
            struct_ser.serialize_field("proposal", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgProposeModelUpdate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "proposer",
            "proposal",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Proposer,
            Proposal,
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
                            "proposer" => Ok(GeneratedField::Proposer),
                            "proposal" => Ok(GeneratedField::Proposal),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgProposeModelUpdate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgProposeModelUpdate")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgProposeModelUpdate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut proposer__ = None;
                let mut proposal__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Proposer => {
                            if proposer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposer"));
                            }
                            proposer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Proposal => {
                            if proposal__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposal"));
                            }
                            proposal__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgProposeModelUpdate {
                    proposer: proposer__.unwrap_or_default(),
                    proposal: proposal__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgProposeModelUpdate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgProposeModelUpdateResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.proposal_id != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgProposeModelUpdateResponse", len)?;
        if self.proposal_id != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("proposalId", ToString::to_string(&self.proposal_id).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgProposeModelUpdateResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "proposal_id",
            "proposalId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProposalId,
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
                            "proposalId" | "proposal_id" => Ok(GeneratedField::ProposalId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgProposeModelUpdateResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgProposeModelUpdateResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgProposeModelUpdateResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut proposal_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProposalId => {
                            if proposal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proposalId"));
                            }
                            proposal_id__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgProposeModelUpdateResponse {
                    proposal_id: proposal_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgProposeModelUpdateResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRebindWallet {
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
        if !self.new_binding_signature.is_empty() {
            len += 1;
        }
        if !self.new_binding_pub_key.is_empty() {
            len += 1;
        }
        if !self.old_signature.is_empty() {
            len += 1;
        }
        if !self.mfa_proof.is_empty() {
            len += 1;
        }
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRebindWallet", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.new_binding_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("newBindingSignature", pbjson::private::base64::encode(&self.new_binding_signature).as_str())?;
        }
        if !self.new_binding_pub_key.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("newBindingPubKey", pbjson::private::base64::encode(&self.new_binding_pub_key).as_str())?;
        }
        if !self.old_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("oldSignature", pbjson::private::base64::encode(&self.old_signature).as_str())?;
        }
        if !self.mfa_proof.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("mfaProof", pbjson::private::base64::encode(&self.mfa_proof).as_str())?;
        }
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRebindWallet {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "new_binding_signature",
            "newBindingSignature",
            "new_binding_pub_key",
            "newBindingPubKey",
            "old_signature",
            "oldSignature",
            "mfa_proof",
            "mfaProof",
            "device_fingerprint",
            "deviceFingerprint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            NewBindingSignature,
            NewBindingPubKey,
            OldSignature,
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
                            "newBindingSignature" | "new_binding_signature" => Ok(GeneratedField::NewBindingSignature),
                            "newBindingPubKey" | "new_binding_pub_key" => Ok(GeneratedField::NewBindingPubKey),
                            "oldSignature" | "old_signature" => Ok(GeneratedField::OldSignature),
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
            type Value = MsgRebindWallet;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRebindWallet")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRebindWallet, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut new_binding_signature__ = None;
                let mut new_binding_pub_key__ = None;
                let mut old_signature__ = None;
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
                        GeneratedField::NewBindingSignature => {
                            if new_binding_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newBindingSignature"));
                            }
                            new_binding_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewBindingPubKey => {
                            if new_binding_pub_key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newBindingPubKey"));
                            }
                            new_binding_pub_key__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::OldSignature => {
                            if old_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("oldSignature"));
                            }
                            old_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MfaProof => {
                            if mfa_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mfaProof"));
                            }
                            mfa_proof__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRebindWallet {
                    sender: sender__.unwrap_or_default(),
                    new_binding_signature: new_binding_signature__.unwrap_or_default(),
                    new_binding_pub_key: new_binding_pub_key__.unwrap_or_default(),
                    old_signature: old_signature__.unwrap_or_default(),
                    mfa_proof: mfa_proof__.unwrap_or_default(),
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRebindWallet", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRebindWalletResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.rebound_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRebindWalletResponse", len)?;
        if self.rebound_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reboundAt", ToString::to_string(&self.rebound_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRebindWalletResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "rebound_at",
            "reboundAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReboundAt,
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
                            "reboundAt" | "rebound_at" => Ok(GeneratedField::ReboundAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRebindWalletResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRebindWalletResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRebindWalletResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut rebound_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReboundAt => {
                            if rebound_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reboundAt"));
                            }
                            rebound_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRebindWalletResponse {
                    rebound_at: rebound_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRebindWalletResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterComplianceProvider {
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
        if self.provider.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRegisterComplianceProvider", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.provider.as_ref() {
            struct_ser.serialize_field("provider", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterComplianceProvider {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "provider",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            Provider,
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
                            "provider" => Ok(GeneratedField::Provider),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterComplianceProvider;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRegisterComplianceProvider")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterComplianceProvider, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut provider__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgRegisterComplianceProvider {
                    authority: authority__.unwrap_or_default(),
                    provider: provider__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRegisterComplianceProvider", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterComplianceProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRegisterComplianceProviderResponse", len)?;
        if !self.provider_id.is_empty() {
            struct_ser.serialize_field("providerId", &self.provider_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterComplianceProviderResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_id",
            "providerId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderId,
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
                            "providerId" | "provider_id" => Ok(GeneratedField::ProviderId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterComplianceProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRegisterComplianceProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterComplianceProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderId => {
                            if provider_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerId"));
                            }
                            provider_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRegisterComplianceProviderResponse {
                    provider_id: provider_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRegisterComplianceProviderResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterModel {
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
        if self.model_info.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRegisterModel", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.model_info.as_ref() {
            struct_ser.serialize_field("modelInfo", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterModel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "model_info",
            "modelInfo",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ModelInfo,
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
                            "modelInfo" | "model_info" => Ok(GeneratedField::ModelInfo),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRegisterModel;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRegisterModel")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterModel, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut model_info__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelInfo => {
                            if model_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelInfo"));
                            }
                            model_info__ = map_.next_value()?;
                        }
                    }
                }
                Ok(MsgRegisterModel {
                    authority: authority__.unwrap_or_default(),
                    model_info: model_info__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRegisterModel", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRegisterModelResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.model_id.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRegisterModelResponse", len)?;
        if !self.model_id.is_empty() {
            struct_ser.serialize_field("modelId", &self.model_id)?;
        }
        if self.status != 0 {
            let v = ModelStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRegisterModelResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "model_id",
            "modelId",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ModelId,
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
                            "modelId" | "model_id" => Ok(GeneratedField::ModelId),
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
            type Value = MsgRegisterModelResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRegisterModelResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRegisterModelResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut model_id__ = None;
                let mut status__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ModelId => {
                            if model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelId"));
                            }
                            model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ModelStatus>()? as i32);
                        }
                    }
                }
                Ok(MsgRegisterModelResponse {
                    model_id: model_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRegisterModelResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgReportModelVersion {
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
        if !self.model_versions.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgReportModelVersion", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.model_versions.is_empty() {
            struct_ser.serialize_field("modelVersions", &self.model_versions)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgReportModelVersion {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "model_versions",
            "modelVersions",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            ModelVersions,
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
                            "modelVersions" | "model_versions" => Ok(GeneratedField::ModelVersions),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgReportModelVersion;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgReportModelVersion")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgReportModelVersion, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut model_versions__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelVersions => {
                            if model_versions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersions"));
                            }
                            model_versions__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                    }
                }
                Ok(MsgReportModelVersion {
                    validator_address: validator_address__.unwrap_or_default(),
                    model_versions: model_versions__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgReportModelVersion", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgReportModelVersionResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.is_synced {
            len += 1;
        }
        if !self.mismatched_models.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgReportModelVersionResponse", len)?;
        if self.is_synced {
            struct_ser.serialize_field("isSynced", &self.is_synced)?;
        }
        if !self.mismatched_models.is_empty() {
            struct_ser.serialize_field("mismatchedModels", &self.mismatched_models)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgReportModelVersionResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "is_synced",
            "isSynced",
            "mismatched_models",
            "mismatchedModels",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            IsSynced,
            MismatchedModels,
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
                            "isSynced" | "is_synced" => Ok(GeneratedField::IsSynced),
                            "mismatchedModels" | "mismatched_models" => Ok(GeneratedField::MismatchedModels),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgReportModelVersionResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgReportModelVersionResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgReportModelVersionResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut is_synced__ = None;
                let mut mismatched_models__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::IsSynced => {
                            if is_synced__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isSynced"));
                            }
                            is_synced__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MismatchedModels => {
                            if mismatched_models__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mismatchedModels"));
                            }
                            mismatched_models__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgReportModelVersionResponse {
                    is_synced: is_synced__.unwrap_or_default(),
                    mismatched_models: mismatched_models__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgReportModelVersionResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRequestVerification {
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
        if !self.scope_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRequestVerification", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRequestVerification {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "scope_id",
            "scopeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ScopeId,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRequestVerification;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRequestVerification")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRequestVerification, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut scope_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRequestVerification {
                    sender: sender__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRequestVerification", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRequestVerificationResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.requested_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRequestVerificationResponse", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.status != 0 {
            let v = VerificationStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.requested_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("requestedAt", ToString::to_string(&self.requested_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRequestVerificationResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "status",
            "requested_at",
            "requestedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            Status,
            RequestedAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "status" => Ok(GeneratedField::Status),
                            "requestedAt" | "requested_at" => Ok(GeneratedField::RequestedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRequestVerificationResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRequestVerificationResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRequestVerificationResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut status__ = None;
                let mut requested_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::RequestedAt => {
                            if requested_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requestedAt"));
                            }
                            requested_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRequestVerificationResponse {
                    scope_id: scope_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    requested_at: requested_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRequestVerificationResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveAppeal {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.resolver.is_empty() {
            len += 1;
        }
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        if self.resolution != 0 {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if self.score_adjustment != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgResolveAppeal", len)?;
        if !self.resolver.is_empty() {
            struct_ser.serialize_field("resolver", &self.resolver)?;
        }
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        if self.resolution != 0 {
            let v = AppealStatus::try_from(self.resolution)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.resolution)))?;
            struct_ser.serialize_field("resolution", &v)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if self.score_adjustment != 0 {
            struct_ser.serialize_field("scoreAdjustment", &self.score_adjustment)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveAppeal {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "resolver",
            "appeal_id",
            "appealId",
            "resolution",
            "reason",
            "score_adjustment",
            "scoreAdjustment",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Resolver,
            AppealId,
            Resolution,
            Reason,
            ScoreAdjustment,
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
                            "resolver" => Ok(GeneratedField::Resolver),
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "reason" => Ok(GeneratedField::Reason),
                            "scoreAdjustment" | "score_adjustment" => Ok(GeneratedField::ScoreAdjustment),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgResolveAppeal;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgResolveAppeal")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveAppeal, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut resolver__ = None;
                let mut appeal_id__ = None;
                let mut resolution__ = None;
                let mut reason__ = None;
                let mut score_adjustment__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Resolver => {
                            if resolver__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolver"));
                            }
                            resolver__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value::<AppealStatus>()? as i32);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScoreAdjustment => {
                            if score_adjustment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreAdjustment"));
                            }
                            score_adjustment__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgResolveAppeal {
                    resolver: resolver__.unwrap_or_default(),
                    appeal_id: appeal_id__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    score_adjustment: score_adjustment__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgResolveAppeal", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgResolveAppealResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        if self.resolution != 0 {
            len += 1;
        }
        if self.resolved_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgResolveAppealResponse", len)?;
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        if self.resolution != 0 {
            let v = AppealStatus::try_from(self.resolution)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.resolution)))?;
            struct_ser.serialize_field("resolution", &v)?;
        }
        if self.resolved_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("resolvedAt", ToString::to_string(&self.resolved_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgResolveAppealResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_id",
            "appealId",
            "resolution",
            "resolved_at",
            "resolvedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealId,
            Resolution,
            ResolvedAt,
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
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            "resolution" => Ok(GeneratedField::Resolution),
                            "resolvedAt" | "resolved_at" => Ok(GeneratedField::ResolvedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgResolveAppealResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgResolveAppealResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgResolveAppealResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_id__ = None;
                let mut resolution__ = None;
                let mut resolved_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resolution => {
                            if resolution__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolution"));
                            }
                            resolution__ = Some(map_.next_value::<AppealStatus>()? as i32);
                        }
                        GeneratedField::ResolvedAt => {
                            if resolved_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resolvedAt"));
                            }
                            resolved_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgResolveAppealResponse {
                    appeal_id: appeal_id__.unwrap_or_default(),
                    resolution: resolution__.unwrap_or_default(),
                    resolved_at: resolved_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgResolveAppealResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeModel {
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
        if !self.model_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRevokeModel", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.model_id.is_empty() {
            struct_ser.serialize_field("modelId", &self.model_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeModel {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "model_id",
            "modelId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ModelId,
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
                            "authority" => Ok(GeneratedField::Authority),
                            "modelId" | "model_id" => Ok(GeneratedField::ModelId),
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
            type Value = MsgRevokeModel;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRevokeModel")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeModel, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut model_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelId => {
                            if model_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelId"));
                            }
                            model_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRevokeModel {
                    authority: authority__.unwrap_or_default(),
                    model_id: model_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRevokeModel", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeModelResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRevokeModelResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeModelResponse {
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
            type Value = MsgRevokeModelResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRevokeModelResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeModelResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgRevokeModelResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRevokeModelResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeScope {
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
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRevokeScope", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeScope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "scope_id",
            "scopeId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ScopeId,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
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
            type Value = MsgRevokeScope;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRevokeScope")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeScope, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut scope_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgRevokeScope {
                    sender: sender__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRevokeScope", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeScopeFromWallet {
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
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.user_signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRevokeScopeFromWallet", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.user_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userSignature", pbjson::private::base64::encode(&self.user_signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeScopeFromWallet {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "scope_id",
            "scopeId",
            "reason",
            "user_signature",
            "userSignature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ScopeId,
            Reason,
            UserSignature,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "reason" => Ok(GeneratedField::Reason),
                            "userSignature" | "user_signature" => Ok(GeneratedField::UserSignature),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeScopeFromWallet;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRevokeScopeFromWallet")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeScopeFromWallet, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut scope_id__ = None;
                let mut reason__ = None;
                let mut user_signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::UserSignature => {
                            if user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userSignature"));
                            }
                            user_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRevokeScopeFromWallet {
                    sender: sender__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    user_signature: user_signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRevokeScopeFromWallet", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeScopeFromWalletResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRevokeScopeFromWalletResponse", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeScopeFromWalletResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "revoked_at",
            "revokedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            RevokedAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeScopeFromWalletResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRevokeScopeFromWalletResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeScopeFromWalletResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut revoked_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRevokeScopeFromWalletResponse {
                    scope_id: scope_id__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRevokeScopeFromWalletResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgRevokeScopeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgRevokeScopeResponse", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgRevokeScopeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "revoked_at",
            "revokedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            RevokedAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgRevokeScopeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgRevokeScopeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgRevokeScopeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut revoked_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgRevokeScopeResponse {
                    scope_id: scope_id__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgRevokeScopeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitAppeal {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.submitter.is_empty() {
            len += 1;
        }
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        if !self.evidence_hashes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgSubmitAppeal", len)?;
        if !self.submitter.is_empty() {
            struct_ser.serialize_field("submitter", &self.submitter)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        if !self.evidence_hashes.is_empty() {
            struct_ser.serialize_field("evidenceHashes", &self.evidence_hashes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitAppeal {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "submitter",
            "scope_id",
            "scopeId",
            "reason",
            "evidence_hashes",
            "evidenceHashes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Submitter,
            ScopeId,
            Reason,
            EvidenceHashes,
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
                            "submitter" => Ok(GeneratedField::Submitter),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "reason" => Ok(GeneratedField::Reason),
                            "evidenceHashes" | "evidence_hashes" => Ok(GeneratedField::EvidenceHashes),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitAppeal;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgSubmitAppeal")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitAppeal, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut submitter__ = None;
                let mut scope_id__ = None;
                let mut reason__ = None;
                let mut evidence_hashes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Submitter => {
                            if submitter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submitter"));
                            }
                            submitter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EvidenceHashes => {
                            if evidence_hashes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("evidenceHashes"));
                            }
                            evidence_hashes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitAppeal {
                    submitter: submitter__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                    evidence_hashes: evidence_hashes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgSubmitAppeal", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitAppealResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.appeal_number != 0 {
            len += 1;
        }
        if self.submitted_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgSubmitAppealResponse", len)?;
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        if self.status != 0 {
            let v = AppealStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.appeal_number != 0 {
            struct_ser.serialize_field("appealNumber", &self.appeal_number)?;
        }
        if self.submitted_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("submittedAt", ToString::to_string(&self.submitted_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitAppealResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_id",
            "appealId",
            "status",
            "appeal_number",
            "appealNumber",
            "submitted_at",
            "submittedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealId,
            Status,
            AppealNumber,
            SubmittedAt,
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
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            "status" => Ok(GeneratedField::Status),
                            "appealNumber" | "appeal_number" => Ok(GeneratedField::AppealNumber),
                            "submittedAt" | "submitted_at" => Ok(GeneratedField::SubmittedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitAppealResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgSubmitAppealResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitAppealResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_id__ = None;
                let mut status__ = None;
                let mut appeal_number__ = None;
                let mut submitted_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<AppealStatus>()? as i32);
                        }
                        GeneratedField::AppealNumber => {
                            if appeal_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealNumber"));
                            }
                            appeal_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SubmittedAt => {
                            if submitted_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submittedAt"));
                            }
                            submitted_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgSubmitAppealResponse {
                    appeal_id: appeal_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    appeal_number: appeal_number__.unwrap_or_default(),
                    submitted_at: submitted_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgSubmitAppealResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitComplianceCheck {
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
        if !self.target_address.is_empty() {
            len += 1;
        }
        if !self.check_results.is_empty() {
            len += 1;
        }
        if !self.provider_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgSubmitComplianceCheck", len)?;
        if !self.provider_address.is_empty() {
            struct_ser.serialize_field("providerAddress", &self.provider_address)?;
        }
        if !self.target_address.is_empty() {
            struct_ser.serialize_field("targetAddress", &self.target_address)?;
        }
        if !self.check_results.is_empty() {
            struct_ser.serialize_field("checkResults", &self.check_results)?;
        }
        if !self.provider_id.is_empty() {
            struct_ser.serialize_field("providerId", &self.provider_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitComplianceCheck {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_address",
            "providerAddress",
            "target_address",
            "targetAddress",
            "check_results",
            "checkResults",
            "provider_id",
            "providerId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderAddress,
            TargetAddress,
            CheckResults,
            ProviderId,
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
                            "targetAddress" | "target_address" => Ok(GeneratedField::TargetAddress),
                            "checkResults" | "check_results" => Ok(GeneratedField::CheckResults),
                            "providerId" | "provider_id" => Ok(GeneratedField::ProviderId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitComplianceCheck;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgSubmitComplianceCheck")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitComplianceCheck, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_address__ = None;
                let mut target_address__ = None;
                let mut check_results__ = None;
                let mut provider_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderAddress => {
                            if provider_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerAddress"));
                            }
                            provider_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TargetAddress => {
                            if target_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("targetAddress"));
                            }
                            target_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::CheckResults => {
                            if check_results__.is_some() {
                                return Err(serde::de::Error::duplicate_field("checkResults"));
                            }
                            check_results__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ProviderId => {
                            if provider_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerId"));
                            }
                            provider_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitComplianceCheck {
                    provider_address: provider_address__.unwrap_or_default(),
                    target_address: target_address__.unwrap_or_default(),
                    check_results: check_results__.unwrap_or_default(),
                    provider_id: provider_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgSubmitComplianceCheck", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitComplianceCheckResponse {
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
        if self.risk_score != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgSubmitComplianceCheckResponse", len)?;
        if self.status != 0 {
            let v = ComplianceStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.risk_score != 0 {
            struct_ser.serialize_field("riskScore", &self.risk_score)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitComplianceCheckResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "status",
            "risk_score",
            "riskScore",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Status,
            RiskScore,
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
                            "riskScore" | "risk_score" => Ok(GeneratedField::RiskScore),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitComplianceCheckResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgSubmitComplianceCheckResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitComplianceCheckResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut status__ = None;
                let mut risk_score__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ComplianceStatus>()? as i32);
                        }
                        GeneratedField::RiskScore => {
                            if risk_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("riskScore"));
                            }
                            risk_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgSubmitComplianceCheckResponse {
                    status: status__.unwrap_or_default(),
                    risk_score: risk_score__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgSubmitComplianceCheckResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateBorderlineParams {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateBorderlineParams", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateBorderlineParams {
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
            type Value = MsgUpdateBorderlineParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateBorderlineParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateBorderlineParams, V::Error>
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
                Ok(MsgUpdateBorderlineParams {
                    authority: authority__.unwrap_or_default(),
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateBorderlineParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateBorderlineParamsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateBorderlineParamsResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateBorderlineParamsResponse {
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
            type Value = MsgUpdateBorderlineParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateBorderlineParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateBorderlineParamsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateBorderlineParamsResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateBorderlineParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateComplianceParams {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateComplianceParams", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateComplianceParams {
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
            type Value = MsgUpdateComplianceParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateComplianceParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateComplianceParams, V::Error>
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
                Ok(MsgUpdateComplianceParams {
                    authority: authority__.unwrap_or_default(),
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateComplianceParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateComplianceParamsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateComplianceParamsResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateComplianceParamsResponse {
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
            type Value = MsgUpdateComplianceParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateComplianceParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateComplianceParamsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgUpdateComplianceParamsResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateComplianceParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateConsentSettings {
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
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.grant_consent {
            len += 1;
        }
        if !self.purpose.is_empty() {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if self.global_settings.is_some() {
            len += 1;
        }
        if !self.user_signature.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateConsentSettings", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.grant_consent {
            struct_ser.serialize_field("grantConsent", &self.grant_consent)?;
        }
        if !self.purpose.is_empty() {
            struct_ser.serialize_field("purpose", &self.purpose)?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if let Some(v) = self.global_settings.as_ref() {
            struct_ser.serialize_field("globalSettings", v)?;
        }
        if !self.user_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userSignature", pbjson::private::base64::encode(&self.user_signature).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateConsentSettings {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "scope_id",
            "scopeId",
            "grant_consent",
            "grantConsent",
            "purpose",
            "expires_at",
            "expiresAt",
            "global_settings",
            "globalSettings",
            "user_signature",
            "userSignature",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ScopeId,
            GrantConsent,
            Purpose,
            ExpiresAt,
            GlobalSettings,
            UserSignature,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "grantConsent" | "grant_consent" => Ok(GeneratedField::GrantConsent),
                            "purpose" => Ok(GeneratedField::Purpose),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "globalSettings" | "global_settings" => Ok(GeneratedField::GlobalSettings),
                            "userSignature" | "user_signature" => Ok(GeneratedField::UserSignature),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateConsentSettings;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateConsentSettings")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateConsentSettings, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut scope_id__ = None;
                let mut grant_consent__ = None;
                let mut purpose__ = None;
                let mut expires_at__ = None;
                let mut global_settings__ = None;
                let mut user_signature__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GrantConsent => {
                            if grant_consent__.is_some() {
                                return Err(serde::de::Error::duplicate_field("grantConsent"));
                            }
                            grant_consent__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Purpose => {
                            if purpose__.is_some() {
                                return Err(serde::de::Error::duplicate_field("purpose"));
                            }
                            purpose__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GlobalSettings => {
                            if global_settings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("globalSettings"));
                            }
                            global_settings__ = map_.next_value()?;
                        }
                        GeneratedField::UserSignature => {
                            if user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userSignature"));
                            }
                            user_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUpdateConsentSettings {
                    sender: sender__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    grant_consent: grant_consent__.unwrap_or_default(),
                    purpose: purpose__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    global_settings: global_settings__,
                    user_signature: user_signature__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateConsentSettings", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateConsentSettingsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.updated_at != 0 {
            len += 1;
        }
        if self.consent_version != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateConsentSettingsResponse", len)?;
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        if self.consent_version != 0 {
            struct_ser.serialize_field("consentVersion", &self.consent_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateConsentSettingsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "updated_at",
            "updatedAt",
            "consent_version",
            "consentVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            UpdatedAt,
            ConsentVersion,
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
                            "updatedAt" | "updated_at" => Ok(GeneratedField::UpdatedAt),
                            "consentVersion" | "consent_version" => Ok(GeneratedField::ConsentVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateConsentSettingsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateConsentSettingsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateConsentSettingsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut updated_at__ = None;
                let mut consent_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ConsentVersion => {
                            if consent_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentVersion"));
                            }
                            consent_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUpdateConsentSettingsResponse {
                    updated_at: updated_at__.unwrap_or_default(),
                    consent_version: consent_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateConsentSettingsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateDerivedFeatures {
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
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.face_embedding_hash.is_empty() {
            len += 1;
        }
        if !self.doc_field_hashes.is_empty() {
            len += 1;
        }
        if !self.biometric_hash.is_empty() {
            len += 1;
        }
        if !self.liveness_proof_hash.is_empty() {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateDerivedFeatures", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.face_embedding_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("faceEmbeddingHash", pbjson::private::base64::encode(&self.face_embedding_hash).as_str())?;
        }
        if !self.doc_field_hashes.is_empty() {
            let v: std::collections::HashMap<_, _> = self.doc_field_hashes.iter()
                .map(|(k, v)| (k, pbjson::private::base64::encode(v))).collect();
            struct_ser.serialize_field("docFieldHashes", &v)?;
        }
        if !self.biometric_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("biometricHash", pbjson::private::base64::encode(&self.biometric_hash).as_str())?;
        }
        if !self.liveness_proof_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("livenessProofHash", pbjson::private::base64::encode(&self.liveness_proof_hash).as_str())?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateDerivedFeatures {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "account_address",
            "accountAddress",
            "face_embedding_hash",
            "faceEmbeddingHash",
            "doc_field_hashes",
            "docFieldHashes",
            "biometric_hash",
            "biometricHash",
            "liveness_proof_hash",
            "livenessProofHash",
            "model_version",
            "modelVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            AccountAddress,
            FaceEmbeddingHash,
            DocFieldHashes,
            BiometricHash,
            LivenessProofHash,
            ModelVersion,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "faceEmbeddingHash" | "face_embedding_hash" => Ok(GeneratedField::FaceEmbeddingHash),
                            "docFieldHashes" | "doc_field_hashes" => Ok(GeneratedField::DocFieldHashes),
                            "biometricHash" | "biometric_hash" => Ok(GeneratedField::BiometricHash),
                            "livenessProofHash" | "liveness_proof_hash" => Ok(GeneratedField::LivenessProofHash),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateDerivedFeatures;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateDerivedFeatures")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateDerivedFeatures, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut account_address__ = None;
                let mut face_embedding_hash__ = None;
                let mut doc_field_hashes__ = None;
                let mut biometric_hash__ = None;
                let mut liveness_proof_hash__ = None;
                let mut model_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FaceEmbeddingHash => {
                            if face_embedding_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("faceEmbeddingHash"));
                            }
                            face_embedding_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DocFieldHashes => {
                            if doc_field_hashes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("docFieldHashes"));
                            }
                            doc_field_hashes__ = Some(
                                map_.next_value::<std::collections::HashMap<_, ::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|(k,v)| (k, v.0)).collect()
                            );
                        }
                        GeneratedField::BiometricHash => {
                            if biometric_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("biometricHash"));
                            }
                            biometric_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LivenessProofHash => {
                            if liveness_proof_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("livenessProofHash"));
                            }
                            liveness_proof_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateDerivedFeatures {
                    sender: sender__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    face_embedding_hash: face_embedding_hash__.unwrap_or_default(),
                    doc_field_hashes: doc_field_hashes__.unwrap_or_default(),
                    biometric_hash: biometric_hash__.unwrap_or_default(),
                    liveness_proof_hash: liveness_proof_hash__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateDerivedFeatures", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateDerivedFeaturesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateDerivedFeaturesResponse", len)?;
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateDerivedFeaturesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "updated_at",
            "updatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = MsgUpdateDerivedFeaturesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateDerivedFeaturesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateDerivedFeaturesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUpdateDerivedFeaturesResponse {
                    updated_at: updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateDerivedFeaturesResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateScore {
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
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.new_score != 0 {
            len += 1;
        }
        if !self.score_version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateScore", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.new_score != 0 {
            struct_ser.serialize_field("newScore", &self.new_score)?;
        }
        if !self.score_version.is_empty() {
            struct_ser.serialize_field("scoreVersion", &self.score_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateScore {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "account_address",
            "accountAddress",
            "new_score",
            "newScore",
            "score_version",
            "scoreVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            AccountAddress,
            NewScore,
            ScoreVersion,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "newScore" | "new_score" => Ok(GeneratedField::NewScore),
                            "scoreVersion" | "score_version" => Ok(GeneratedField::ScoreVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUpdateScore;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateScore")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateScore, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut account_address__ = None;
                let mut new_score__ = None;
                let mut score_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewScore => {
                            if new_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newScore"));
                            }
                            new_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ScoreVersion => {
                            if score_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scoreVersion"));
                            }
                            score_version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateScore {
                    sender: sender__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    new_score: new_score__.unwrap_or_default(),
                    score_version: score_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateScore", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateScoreResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.previous_score != 0 {
            len += 1;
        }
        if self.new_score != 0 {
            len += 1;
        }
        if self.previous_tier != 0 {
            len += 1;
        }
        if self.new_tier != 0 {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateScoreResponse", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.previous_score != 0 {
            struct_ser.serialize_field("previousScore", &self.previous_score)?;
        }
        if self.new_score != 0 {
            struct_ser.serialize_field("newScore", &self.new_score)?;
        }
        if self.previous_tier != 0 {
            let v = IdentityTier::try_from(self.previous_tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_tier)))?;
            struct_ser.serialize_field("previousTier", &v)?;
        }
        if self.new_tier != 0 {
            let v = IdentityTier::try_from(self.new_tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_tier)))?;
            struct_ser.serialize_field("newTier", &v)?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateScoreResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "previous_score",
            "previousScore",
            "new_score",
            "newScore",
            "previous_tier",
            "previousTier",
            "new_tier",
            "newTier",
            "updated_at",
            "updatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            PreviousScore,
            NewScore,
            PreviousTier,
            NewTier,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "previousScore" | "previous_score" => Ok(GeneratedField::PreviousScore),
                            "newScore" | "new_score" => Ok(GeneratedField::NewScore),
                            "previousTier" | "previous_tier" => Ok(GeneratedField::PreviousTier),
                            "newTier" | "new_tier" => Ok(GeneratedField::NewTier),
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
            type Value = MsgUpdateScoreResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateScoreResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateScoreResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut previous_score__ = None;
                let mut new_score__ = None;
                let mut previous_tier__ = None;
                let mut new_tier__ = None;
                let mut updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PreviousScore => {
                            if previous_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousScore"));
                            }
                            previous_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewScore => {
                            if new_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newScore"));
                            }
                            new_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousTier => {
                            if previous_tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousTier"));
                            }
                            previous_tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::NewTier => {
                            if new_tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newTier"));
                            }
                            new_tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUpdateScoreResponse {
                    account_address: account_address__.unwrap_or_default(),
                    previous_score: previous_score__.unwrap_or_default(),
                    new_score: new_score__.unwrap_or_default(),
                    previous_tier: previous_tier__.unwrap_or_default(),
                    new_tier: new_tier__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateScoreResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateVerificationStatus {
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
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateVerificationStatus", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.new_status != 0 {
            let v = VerificationStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateVerificationStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "account_address",
            "accountAddress",
            "scope_id",
            "scopeId",
            "new_status",
            "newStatus",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            AccountAddress,
            ScopeId,
            NewStatus,
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
                            "sender" => Ok(GeneratedField::Sender),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
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
            type Value = MsgUpdateVerificationStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateVerificationStatus")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateVerificationStatus, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut account_address__ = None;
                let mut scope_id__ = None;
                let mut new_status__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUpdateVerificationStatus {
                    sender: sender__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateVerificationStatus", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUpdateVerificationStatusResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.previous_status != 0 {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if self.updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUpdateVerificationStatusResponse", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.previous_status != 0 {
            let v = VerificationStatus::try_from(self.previous_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_status)))?;
            struct_ser.serialize_field("previousStatus", &v)?;
        }
        if self.new_status != 0 {
            let v = VerificationStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if self.updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("updatedAt", ToString::to_string(&self.updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUpdateVerificationStatusResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "previous_status",
            "previousStatus",
            "new_status",
            "newStatus",
            "updated_at",
            "updatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            PreviousStatus,
            NewStatus,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "previousStatus" | "previous_status" => Ok(GeneratedField::PreviousStatus),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
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
            type Value = MsgUpdateVerificationStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUpdateVerificationStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUpdateVerificationStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut previous_status__ = None;
                let mut new_status__ = None;
                let mut updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::PreviousStatus => {
                            if previous_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousStatus"));
                            }
                            previous_status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::UpdatedAt => {
                            if updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("updatedAt"));
                            }
                            updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUpdateVerificationStatusResponse {
                    scope_id: scope_id__.unwrap_or_default(),
                    previous_status: previous_status__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    updated_at: updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUpdateVerificationStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUploadScope {
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
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if self.encrypted_payload.is_some() {
            len += 1;
        }
        if !self.salt.is_empty() {
            len += 1;
        }
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        if !self.client_id.is_empty() {
            len += 1;
        }
        if !self.client_signature.is_empty() {
            len += 1;
        }
        if !self.user_signature.is_empty() {
            len += 1;
        }
        if !self.payload_hash.is_empty() {
            len += 1;
        }
        if self.capture_timestamp != 0 {
            len += 1;
        }
        if !self.geo_hint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUploadScope", len)?;
        if !self.sender.is_empty() {
            struct_ser.serialize_field("sender", &self.sender)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if let Some(v) = self.encrypted_payload.as_ref() {
            struct_ser.serialize_field("encryptedPayload", v)?;
        }
        if !self.salt.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("salt", pbjson::private::base64::encode(&self.salt).as_str())?;
        }
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        if !self.client_id.is_empty() {
            struct_ser.serialize_field("clientId", &self.client_id)?;
        }
        if !self.client_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("clientSignature", pbjson::private::base64::encode(&self.client_signature).as_str())?;
        }
        if !self.user_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userSignature", pbjson::private::base64::encode(&self.user_signature).as_str())?;
        }
        if !self.payload_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("payloadHash", pbjson::private::base64::encode(&self.payload_hash).as_str())?;
        }
        if self.capture_timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("captureTimestamp", ToString::to_string(&self.capture_timestamp).as_str())?;
        }
        if !self.geo_hint.is_empty() {
            struct_ser.serialize_field("geoHint", &self.geo_hint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUploadScope {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "sender",
            "scope_id",
            "scopeId",
            "scope_type",
            "scopeType",
            "encrypted_payload",
            "encryptedPayload",
            "salt",
            "device_fingerprint",
            "deviceFingerprint",
            "client_id",
            "clientId",
            "client_signature",
            "clientSignature",
            "user_signature",
            "userSignature",
            "payload_hash",
            "payloadHash",
            "capture_timestamp",
            "captureTimestamp",
            "geo_hint",
            "geoHint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Sender,
            ScopeId,
            ScopeType,
            EncryptedPayload,
            Salt,
            DeviceFingerprint,
            ClientId,
            ClientSignature,
            UserSignature,
            PayloadHash,
            CaptureTimestamp,
            GeoHint,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "encryptedPayload" | "encrypted_payload" => Ok(GeneratedField::EncryptedPayload),
                            "salt" => Ok(GeneratedField::Salt),
                            "deviceFingerprint" | "device_fingerprint" => Ok(GeneratedField::DeviceFingerprint),
                            "clientId" | "client_id" => Ok(GeneratedField::ClientId),
                            "clientSignature" | "client_signature" => Ok(GeneratedField::ClientSignature),
                            "userSignature" | "user_signature" => Ok(GeneratedField::UserSignature),
                            "payloadHash" | "payload_hash" => Ok(GeneratedField::PayloadHash),
                            "captureTimestamp" | "capture_timestamp" => Ok(GeneratedField::CaptureTimestamp),
                            "geoHint" | "geo_hint" => Ok(GeneratedField::GeoHint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUploadScope;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUploadScope")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUploadScope, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut sender__ = None;
                let mut scope_id__ = None;
                let mut scope_type__ = None;
                let mut encrypted_payload__ = None;
                let mut salt__ = None;
                let mut device_fingerprint__ = None;
                let mut client_id__ = None;
                let mut client_signature__ = None;
                let mut user_signature__ = None;
                let mut payload_hash__ = None;
                let mut capture_timestamp__ = None;
                let mut geo_hint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Sender => {
                            if sender__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sender"));
                            }
                            sender__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::EncryptedPayload => {
                            if encrypted_payload__.is_some() {
                                return Err(serde::de::Error::duplicate_field("encryptedPayload"));
                            }
                            encrypted_payload__ = map_.next_value()?;
                        }
                        GeneratedField::Salt => {
                            if salt__.is_some() {
                                return Err(serde::de::Error::duplicate_field("salt"));
                            }
                            salt__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClientId => {
                            if client_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientId"));
                            }
                            client_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClientSignature => {
                            if client_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientSignature"));
                            }
                            client_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UserSignature => {
                            if user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userSignature"));
                            }
                            user_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PayloadHash => {
                            if payload_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payloadHash"));
                            }
                            payload_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CaptureTimestamp => {
                            if capture_timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("captureTimestamp"));
                            }
                            capture_timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GeoHint => {
                            if geo_hint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("geoHint"));
                            }
                            geo_hint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgUploadScope {
                    sender: sender__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    encrypted_payload: encrypted_payload__,
                    salt: salt__.unwrap_or_default(),
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                    client_id: client_id__.unwrap_or_default(),
                    client_signature: client_signature__.unwrap_or_default(),
                    user_signature: user_signature__.unwrap_or_default(),
                    payload_hash: payload_hash__.unwrap_or_default(),
                    capture_timestamp: capture_timestamp__.unwrap_or_default(),
                    geo_hint: geo_hint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUploadScope", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgUploadScopeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.uploaded_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgUploadScopeResponse", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.status != 0 {
            let v = VerificationStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.uploaded_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("uploadedAt", ToString::to_string(&self.uploaded_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgUploadScopeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "status",
            "uploaded_at",
            "uploadedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            Status,
            UploadedAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "status" => Ok(GeneratedField::Status),
                            "uploadedAt" | "uploaded_at" => Ok(GeneratedField::UploadedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgUploadScopeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgUploadScopeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgUploadScopeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut status__ = None;
                let mut uploaded_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::UploadedAt => {
                            if uploaded_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uploadedAt"));
                            }
                            uploaded_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgUploadScopeResponse {
                    scope_id: scope_id__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    uploaded_at: uploaded_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgUploadScopeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgWithdrawAppeal {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.submitter.is_empty() {
            len += 1;
        }
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgWithdrawAppeal", len)?;
        if !self.submitter.is_empty() {
            struct_ser.serialize_field("submitter", &self.submitter)?;
        }
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgWithdrawAppeal {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "submitter",
            "appeal_id",
            "appealId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Submitter,
            AppealId,
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
                            "submitter" => Ok(GeneratedField::Submitter),
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgWithdrawAppeal;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgWithdrawAppeal")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgWithdrawAppeal, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut submitter__ = None;
                let mut appeal_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Submitter => {
                            if submitter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submitter"));
                            }
                            submitter__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgWithdrawAppeal {
                    submitter: submitter__.unwrap_or_default(),
                    appeal_id: appeal_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgWithdrawAppeal", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgWithdrawAppealResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        if self.withdrawn_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.MsgWithdrawAppealResponse", len)?;
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        if self.withdrawn_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("withdrawnAt", ToString::to_string(&self.withdrawn_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgWithdrawAppealResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_id",
            "appealId",
            "withdrawn_at",
            "withdrawnAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealId,
            WithdrawnAt,
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
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            "withdrawnAt" | "withdrawn_at" => Ok(GeneratedField::WithdrawnAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgWithdrawAppealResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.MsgWithdrawAppealResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgWithdrawAppealResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_id__ = None;
                let mut withdrawn_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::WithdrawnAt => {
                            if withdrawn_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("withdrawnAt"));
                            }
                            withdrawn_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(MsgWithdrawAppealResponse {
                    appeal_id: appeal_id__.unwrap_or_default(),
                    withdrawn_at: withdrawn_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.MsgWithdrawAppealResponse", FIELDS, GeneratedVisitor)
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
        if self.max_scopes_per_account != 0 {
            len += 1;
        }
        if self.max_scopes_per_type != 0 {
            len += 1;
        }
        if self.salt_min_bytes != 0 {
            len += 1;
        }
        if self.salt_max_bytes != 0 {
            len += 1;
        }
        if self.require_client_signature {
            len += 1;
        }
        if self.require_user_signature {
            len += 1;
        }
        if self.verification_expiry_days != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.Params", len)?;
        if self.max_scopes_per_account != 0 {
            struct_ser.serialize_field("maxScopesPerAccount", &self.max_scopes_per_account)?;
        }
        if self.max_scopes_per_type != 0 {
            struct_ser.serialize_field("maxScopesPerType", &self.max_scopes_per_type)?;
        }
        if self.salt_min_bytes != 0 {
            struct_ser.serialize_field("saltMinBytes", &self.salt_min_bytes)?;
        }
        if self.salt_max_bytes != 0 {
            struct_ser.serialize_field("saltMaxBytes", &self.salt_max_bytes)?;
        }
        if self.require_client_signature {
            struct_ser.serialize_field("requireClientSignature", &self.require_client_signature)?;
        }
        if self.require_user_signature {
            struct_ser.serialize_field("requireUserSignature", &self.require_user_signature)?;
        }
        if self.verification_expiry_days != 0 {
            struct_ser.serialize_field("verificationExpiryDays", &self.verification_expiry_days)?;
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
            "max_scopes_per_account",
            "maxScopesPerAccount",
            "max_scopes_per_type",
            "maxScopesPerType",
            "salt_min_bytes",
            "saltMinBytes",
            "salt_max_bytes",
            "saltMaxBytes",
            "require_client_signature",
            "requireClientSignature",
            "require_user_signature",
            "requireUserSignature",
            "verification_expiry_days",
            "verificationExpiryDays",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MaxScopesPerAccount,
            MaxScopesPerType,
            SaltMinBytes,
            SaltMaxBytes,
            RequireClientSignature,
            RequireUserSignature,
            VerificationExpiryDays,
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
                            "maxScopesPerAccount" | "max_scopes_per_account" => Ok(GeneratedField::MaxScopesPerAccount),
                            "maxScopesPerType" | "max_scopes_per_type" => Ok(GeneratedField::MaxScopesPerType),
                            "saltMinBytes" | "salt_min_bytes" => Ok(GeneratedField::SaltMinBytes),
                            "saltMaxBytes" | "salt_max_bytes" => Ok(GeneratedField::SaltMaxBytes),
                            "requireClientSignature" | "require_client_signature" => Ok(GeneratedField::RequireClientSignature),
                            "requireUserSignature" | "require_user_signature" => Ok(GeneratedField::RequireUserSignature),
                            "verificationExpiryDays" | "verification_expiry_days" => Ok(GeneratedField::VerificationExpiryDays),
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
                formatter.write_str("struct virtengine.veid.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut max_scopes_per_account__ = None;
                let mut max_scopes_per_type__ = None;
                let mut salt_min_bytes__ = None;
                let mut salt_max_bytes__ = None;
                let mut require_client_signature__ = None;
                let mut require_user_signature__ = None;
                let mut verification_expiry_days__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MaxScopesPerAccount => {
                            if max_scopes_per_account__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxScopesPerAccount"));
                            }
                            max_scopes_per_account__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxScopesPerType => {
                            if max_scopes_per_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxScopesPerType"));
                            }
                            max_scopes_per_type__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SaltMinBytes => {
                            if salt_min_bytes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("saltMinBytes"));
                            }
                            salt_min_bytes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SaltMaxBytes => {
                            if salt_max_bytes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("saltMaxBytes"));
                            }
                            salt_max_bytes__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequireClientSignature => {
                            if require_client_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireClientSignature"));
                            }
                            require_client_signature__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RequireUserSignature => {
                            if require_user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireUserSignature"));
                            }
                            require_user_signature__ = Some(map_.next_value()?);
                        }
                        GeneratedField::VerificationExpiryDays => {
                            if verification_expiry_days__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verificationExpiryDays"));
                            }
                            verification_expiry_days__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Params {
                    max_scopes_per_account: max_scopes_per_account__.unwrap_or_default(),
                    max_scopes_per_type: max_scopes_per_type__.unwrap_or_default(),
                    salt_min_bytes: salt_min_bytes__.unwrap_or_default(),
                    salt_max_bytes: salt_max_bytes__.unwrap_or_default(),
                    require_client_signature: require_client_signature__.unwrap_or_default(),
                    require_user_signature: require_user_signature__.unwrap_or_default(),
                    verification_expiry_days: verification_expiry_days__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PublicConsentInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.granted {
            len += 1;
        }
        if self.is_active {
            len += 1;
        }
        if !self.purpose.is_empty() {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.PublicConsentInfo", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.granted {
            struct_ser.serialize_field("granted", &self.granted)?;
        }
        if self.is_active {
            struct_ser.serialize_field("isActive", &self.is_active)?;
        }
        if !self.purpose.is_empty() {
            struct_ser.serialize_field("purpose", &self.purpose)?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PublicConsentInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "granted",
            "is_active",
            "isActive",
            "purpose",
            "expires_at",
            "expiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            Granted,
            IsActive,
            Purpose,
            ExpiresAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "granted" => Ok(GeneratedField::Granted),
                            "isActive" | "is_active" => Ok(GeneratedField::IsActive),
                            "purpose" => Ok(GeneratedField::Purpose),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PublicConsentInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.PublicConsentInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PublicConsentInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut granted__ = None;
                let mut is_active__ = None;
                let mut purpose__ = None;
                let mut expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Granted => {
                            if granted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("granted"));
                            }
                            granted__ = Some(map_.next_value()?);
                        }
                        GeneratedField::IsActive => {
                            if is_active__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isActive"));
                            }
                            is_active__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Purpose => {
                            if purpose__.is_some() {
                                return Err(serde::de::Error::duplicate_field("purpose"));
                            }
                            purpose__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(PublicConsentInfo {
                    scope_id: scope_id__.unwrap_or_default(),
                    granted: granted__.unwrap_or_default(),
                    is_active: is_active__.unwrap_or_default(),
                    purpose: purpose__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.PublicConsentInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PublicDerivedFeaturesInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.has_face_embedding {
            len += 1;
        }
        if self.has_biometric {
            len += 1;
        }
        if self.has_liveness_proof {
            len += 1;
        }
        if !self.doc_field_keys.is_empty() {
            len += 1;
        }
        if self.last_computed_at != 0 {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        if self.feature_version != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.PublicDerivedFeaturesInfo", len)?;
        if self.has_face_embedding {
            struct_ser.serialize_field("hasFaceEmbedding", &self.has_face_embedding)?;
        }
        if self.has_biometric {
            struct_ser.serialize_field("hasBiometric", &self.has_biometric)?;
        }
        if self.has_liveness_proof {
            struct_ser.serialize_field("hasLivenessProof", &self.has_liveness_proof)?;
        }
        if !self.doc_field_keys.is_empty() {
            struct_ser.serialize_field("docFieldKeys", &self.doc_field_keys)?;
        }
        if self.last_computed_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastComputedAt", ToString::to_string(&self.last_computed_at).as_str())?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        if self.feature_version != 0 {
            struct_ser.serialize_field("featureVersion", &self.feature_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PublicDerivedFeaturesInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "has_face_embedding",
            "hasFaceEmbedding",
            "has_biometric",
            "hasBiometric",
            "has_liveness_proof",
            "hasLivenessProof",
            "doc_field_keys",
            "docFieldKeys",
            "last_computed_at",
            "lastComputedAt",
            "model_version",
            "modelVersion",
            "feature_version",
            "featureVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            HasFaceEmbedding,
            HasBiometric,
            HasLivenessProof,
            DocFieldKeys,
            LastComputedAt,
            ModelVersion,
            FeatureVersion,
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
                            "hasFaceEmbedding" | "has_face_embedding" => Ok(GeneratedField::HasFaceEmbedding),
                            "hasBiometric" | "has_biometric" => Ok(GeneratedField::HasBiometric),
                            "hasLivenessProof" | "has_liveness_proof" => Ok(GeneratedField::HasLivenessProof),
                            "docFieldKeys" | "doc_field_keys" => Ok(GeneratedField::DocFieldKeys),
                            "lastComputedAt" | "last_computed_at" => Ok(GeneratedField::LastComputedAt),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            "featureVersion" | "feature_version" => Ok(GeneratedField::FeatureVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PublicDerivedFeaturesInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.PublicDerivedFeaturesInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PublicDerivedFeaturesInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut has_face_embedding__ = None;
                let mut has_biometric__ = None;
                let mut has_liveness_proof__ = None;
                let mut doc_field_keys__ = None;
                let mut last_computed_at__ = None;
                let mut model_version__ = None;
                let mut feature_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::HasFaceEmbedding => {
                            if has_face_embedding__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hasFaceEmbedding"));
                            }
                            has_face_embedding__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HasBiometric => {
                            if has_biometric__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hasBiometric"));
                            }
                            has_biometric__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HasLivenessProof => {
                            if has_liveness_proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hasLivenessProof"));
                            }
                            has_liveness_proof__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DocFieldKeys => {
                            if doc_field_keys__.is_some() {
                                return Err(serde::de::Error::duplicate_field("docFieldKeys"));
                            }
                            doc_field_keys__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastComputedAt => {
                            if last_computed_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastComputedAt"));
                            }
                            last_computed_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FeatureVersion => {
                            if feature_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("featureVersion"));
                            }
                            feature_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(PublicDerivedFeaturesInfo {
                    has_face_embedding: has_face_embedding__.unwrap_or_default(),
                    has_biometric: has_biometric__.unwrap_or_default(),
                    has_liveness_proof: has_liveness_proof__.unwrap_or_default(),
                    doc_field_keys: doc_field_keys__.unwrap_or_default(),
                    last_computed_at: last_computed_at__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                    feature_version: feature_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.PublicDerivedFeaturesInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PublicVerificationHistoryEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.entry_id.is_empty() {
            len += 1;
        }
        if self.timestamp != 0 {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.previous_score != 0 {
            len += 1;
        }
        if self.new_score != 0 {
            len += 1;
        }
        if self.previous_status != 0 {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if self.scope_count != 0 {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.PublicVerificationHistoryEntry", len)?;
        if !self.entry_id.is_empty() {
            struct_ser.serialize_field("entryId", &self.entry_id)?;
        }
        if self.timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("timestamp", ToString::to_string(&self.timestamp).as_str())?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if self.previous_score != 0 {
            struct_ser.serialize_field("previousScore", &self.previous_score)?;
        }
        if self.new_score != 0 {
            struct_ser.serialize_field("newScore", &self.new_score)?;
        }
        if self.previous_status != 0 {
            let v = AccountStatus::try_from(self.previous_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_status)))?;
            struct_ser.serialize_field("previousStatus", &v)?;
        }
        if self.new_status != 0 {
            let v = AccountStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if self.scope_count != 0 {
            struct_ser.serialize_field("scopeCount", &self.scope_count)?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PublicVerificationHistoryEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "entry_id",
            "entryId",
            "timestamp",
            "block_height",
            "blockHeight",
            "previous_score",
            "previousScore",
            "new_score",
            "newScore",
            "previous_status",
            "previousStatus",
            "new_status",
            "newStatus",
            "scope_count",
            "scopeCount",
            "model_version",
            "modelVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EntryId,
            Timestamp,
            BlockHeight,
            PreviousScore,
            NewScore,
            PreviousStatus,
            NewStatus,
            ScopeCount,
            ModelVersion,
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
                            "entryId" | "entry_id" => Ok(GeneratedField::EntryId),
                            "timestamp" => Ok(GeneratedField::Timestamp),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "previousScore" | "previous_score" => Ok(GeneratedField::PreviousScore),
                            "newScore" | "new_score" => Ok(GeneratedField::NewScore),
                            "previousStatus" | "previous_status" => Ok(GeneratedField::PreviousStatus),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
                            "scopeCount" | "scope_count" => Ok(GeneratedField::ScopeCount),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PublicVerificationHistoryEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.PublicVerificationHistoryEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PublicVerificationHistoryEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut entry_id__ = None;
                let mut timestamp__ = None;
                let mut block_height__ = None;
                let mut previous_score__ = None;
                let mut new_score__ = None;
                let mut previous_status__ = None;
                let mut new_status__ = None;
                let mut scope_count__ = None;
                let mut model_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EntryId => {
                            if entry_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("entryId"));
                            }
                            entry_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousScore => {
                            if previous_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousScore"));
                            }
                            previous_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewScore => {
                            if new_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newScore"));
                            }
                            new_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousStatus => {
                            if previous_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousStatus"));
                            }
                            previous_status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::ScopeCount => {
                            if scope_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeCount"));
                            }
                            scope_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(PublicVerificationHistoryEntry {
                    entry_id: entry_id__.unwrap_or_default(),
                    timestamp: timestamp__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    previous_score: previous_score__.unwrap_or_default(),
                    new_score: new_score__.unwrap_or_default(),
                    previous_status: previous_status__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    scope_count: scope_count__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.PublicVerificationHistoryEntry", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for PublicWalletInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.wallet_id.is_empty() {
            len += 1;
        }
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.score != 0 {
            len += 1;
        }
        if self.tier != 0 {
            len += 1;
        }
        if self.scope_count != 0 {
            len += 1;
        }
        if self.verified_scope_count != 0 {
            len += 1;
        }
        if self.created_at != 0 {
            len += 1;
        }
        if self.last_updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.PublicWalletInfo", len)?;
        if !self.wallet_id.is_empty() {
            struct_ser.serialize_field("walletId", &self.wallet_id)?;
        }
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.status != 0 {
            let v = WalletStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.score != 0 {
            struct_ser.serialize_field("score", &self.score)?;
        }
        if self.tier != 0 {
            let v = IdentityTier::try_from(self.tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tier)))?;
            struct_ser.serialize_field("tier", &v)?;
        }
        if self.scope_count != 0 {
            struct_ser.serialize_field("scopeCount", &self.scope_count)?;
        }
        if self.verified_scope_count != 0 {
            struct_ser.serialize_field("verifiedScopeCount", &self.verified_scope_count)?;
        }
        if self.created_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("createdAt", ToString::to_string(&self.created_at).as_str())?;
        }
        if self.last_updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdatedAt", ToString::to_string(&self.last_updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for PublicWalletInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "wallet_id",
            "walletId",
            "account_address",
            "accountAddress",
            "status",
            "score",
            "tier",
            "scope_count",
            "scopeCount",
            "verified_scope_count",
            "verifiedScopeCount",
            "created_at",
            "createdAt",
            "last_updated_at",
            "lastUpdatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            WalletId,
            AccountAddress,
            Status,
            Score,
            Tier,
            ScopeCount,
            VerifiedScopeCount,
            CreatedAt,
            LastUpdatedAt,
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
                            "walletId" | "wallet_id" => Ok(GeneratedField::WalletId),
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "status" => Ok(GeneratedField::Status),
                            "score" => Ok(GeneratedField::Score),
                            "tier" => Ok(GeneratedField::Tier),
                            "scopeCount" | "scope_count" => Ok(GeneratedField::ScopeCount),
                            "verifiedScopeCount" | "verified_scope_count" => Ok(GeneratedField::VerifiedScopeCount),
                            "createdAt" | "created_at" => Ok(GeneratedField::CreatedAt),
                            "lastUpdatedAt" | "last_updated_at" => Ok(GeneratedField::LastUpdatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = PublicWalletInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.PublicWalletInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<PublicWalletInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut wallet_id__ = None;
                let mut account_address__ = None;
                let mut status__ = None;
                let mut score__ = None;
                let mut tier__ = None;
                let mut scope_count__ = None;
                let mut verified_scope_count__ = None;
                let mut created_at__ = None;
                let mut last_updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::WalletId => {
                            if wallet_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("walletId"));
                            }
                            wallet_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<WalletStatus>()? as i32);
                        }
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Tier => {
                            if tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tier"));
                            }
                            tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::ScopeCount => {
                            if scope_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeCount"));
                            }
                            scope_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VerifiedScopeCount => {
                            if verified_scope_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedScopeCount"));
                            }
                            verified_scope_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CreatedAt => {
                            if created_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("createdAt"));
                            }
                            created_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastUpdatedAt => {
                            if last_updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdatedAt"));
                            }
                            last_updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(PublicWalletInfo {
                    wallet_id: wallet_id__.unwrap_or_default(),
                    account_address: account_address__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    score: score__.unwrap_or_default(),
                    tier: tier__.unwrap_or_default(),
                    scope_count: scope_count__.unwrap_or_default(),
                    verified_scope_count: verified_scope_count__.unwrap_or_default(),
                    created_at: created_at__.unwrap_or_default(),
                    last_updated_at: last_updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.PublicWalletInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryActiveModelsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryActiveModelsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryActiveModelsRequest {
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
            type Value = QueryActiveModelsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryActiveModelsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryActiveModelsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryActiveModelsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryActiveModelsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryActiveModelsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.state.is_some() {
            len += 1;
        }
        if !self.models.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryActiveModelsResponse", len)?;
        if let Some(v) = self.state.as_ref() {
            struct_ser.serialize_field("state", v)?;
        }
        if !self.models.is_empty() {
            struct_ser.serialize_field("models", &self.models)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryActiveModelsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "models",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            Models,
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
                            "models" => Ok(GeneratedField::Models),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryActiveModelsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryActiveModelsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryActiveModelsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut models__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = map_.next_value()?;
                        }
                        GeneratedField::Models => {
                            if models__.is_some() {
                                return Err(serde::de::Error::duplicate_field("models"));
                            }
                            models__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryActiveModelsResponse {
                    state: state__,
                    models: models__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryActiveModelsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealParamsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealParamsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealParamsRequest {
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
            type Value = QueryAppealParamsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealParamsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealParamsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryAppealParamsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealParamsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealParamsResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealParamsResponse", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealParamsResponse {
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
            type Value = QueryAppealParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealParamsResponse, V::Error>
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
                Ok(QueryAppealParamsResponse {
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeal_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealRequest", len)?;
        if !self.appeal_id.is_empty() {
            struct_ser.serialize_field("appealId", &self.appeal_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal_id",
            "appealId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AppealId,
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
                            "appealId" | "appeal_id" => Ok(GeneratedField::AppealId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryAppealRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AppealId => {
                            if appeal_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appealId"));
                            }
                            appeal_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAppealRequest {
                    appeal_id: appeal_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.appeal.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealResponse", len)?;
        if let Some(v) = self.appeal.as_ref() {
            struct_ser.serialize_field("appeal", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeal",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Appeal,
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
                            "appeal" => Ok(GeneratedField::Appeal),
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
            type Value = QueryAppealResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeal__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Appeal => {
                            if appeal__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appeal"));
                            }
                            appeal__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryAppealResponse {
                    appeal: appeal__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealsByScopeRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealsByScopeRequest", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealsByScopeRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
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
            type Value = QueryAppealsByScopeRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealsByScopeRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealsByScopeRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAppealsByScopeRequest {
                    scope_id: scope_id__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealsByScopeRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealsByScopeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeals.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealsByScopeResponse", len)?;
        if !self.appeals.is_empty() {
            struct_ser.serialize_field("appeals", &self.appeals)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealsByScopeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeals",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Appeals,
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
                            "appeals" => Ok(GeneratedField::Appeals),
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
            type Value = QueryAppealsByScopeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealsByScopeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealsByScopeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeals__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Appeals => {
                            if appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appeals"));
                            }
                            appeals__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAppealsByScopeResponse {
                    appeals: appeals__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealsByScopeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealsRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
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
            type Value = QueryAppealsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAppealsRequest {
                    account_address: account_address__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryAppealsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.appeals.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryAppealsResponse", len)?;
        if !self.appeals.is_empty() {
            struct_ser.serialize_field("appeals", &self.appeals)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryAppealsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "appeals",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Appeals,
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
                            "appeals" => Ok(GeneratedField::Appeals),
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
            type Value = QueryAppealsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryAppealsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryAppealsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut appeals__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Appeals => {
                            if appeals__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appeals"));
                            }
                            appeals__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryAppealsResponse {
                    appeals: appeals__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryAppealsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryApprovedClientsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.active_only {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryApprovedClientsRequest", len)?;
        if self.active_only {
            struct_ser.serialize_field("activeOnly", &self.active_only)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryApprovedClientsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "active_only",
            "activeOnly",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ActiveOnly,
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
                            "activeOnly" | "active_only" => Ok(GeneratedField::ActiveOnly),
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
            type Value = QueryApprovedClientsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryApprovedClientsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryApprovedClientsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut active_only__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ActiveOnly => {
                            if active_only__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activeOnly"));
                            }
                            active_only__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryApprovedClientsRequest {
                    active_only: active_only__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryApprovedClientsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryApprovedClientsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.clients.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryApprovedClientsResponse", len)?;
        if !self.clients.is_empty() {
            struct_ser.serialize_field("clients", &self.clients)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryApprovedClientsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "clients",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Clients,
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
                            "clients" => Ok(GeneratedField::Clients),
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
            type Value = QueryApprovedClientsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryApprovedClientsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryApprovedClientsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut clients__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Clients => {
                            if clients__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clients"));
                            }
                            clients__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryApprovedClientsResponse {
                    clients: clients__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryApprovedClientsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryBorderlineParamsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryBorderlineParamsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryBorderlineParamsRequest {
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
            type Value = QueryBorderlineParamsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryBorderlineParamsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryBorderlineParamsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryBorderlineParamsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryBorderlineParamsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryBorderlineParamsResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryBorderlineParamsResponse", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryBorderlineParamsResponse {
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
            type Value = QueryBorderlineParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryBorderlineParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryBorderlineParamsResponse, V::Error>
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
                Ok(QueryBorderlineParamsResponse {
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryBorderlineParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceParamsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceParamsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceParamsRequest {
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
            type Value = QueryComplianceParamsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceParamsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceParamsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryComplianceParamsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceParamsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceParamsResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceParamsResponse", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceParamsResponse {
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
            type Value = QueryComplianceParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceParamsResponse, V::Error>
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
                Ok(QueryComplianceParamsResponse {
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceProviderRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.provider_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceProviderRequest", len)?;
        if !self.provider_id.is_empty() {
            struct_ser.serialize_field("providerId", &self.provider_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceProviderRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider_id",
            "providerId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProviderId,
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
                            "providerId" | "provider_id" => Ok(GeneratedField::ProviderId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryComplianceProviderRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceProviderRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceProviderRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ProviderId => {
                            if provider_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("providerId"));
                            }
                            provider_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryComplianceProviderRequest {
                    provider_id: provider_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceProviderRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceProviderResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.provider.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceProviderResponse", len)?;
        if let Some(v) = self.provider.as_ref() {
            struct_ser.serialize_field("provider", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceProviderResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "provider",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Provider,
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
                            "provider" => Ok(GeneratedField::Provider),
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
            type Value = QueryComplianceProviderResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceProviderResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceProviderResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut provider__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Provider => {
                            if provider__.is_some() {
                                return Err(serde::de::Error::duplicate_field("provider"));
                            }
                            provider__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryComplianceProviderResponse {
                    provider: provider__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceProviderResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceProvidersRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.active_only {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceProvidersRequest", len)?;
        if self.active_only {
            struct_ser.serialize_field("activeOnly", &self.active_only)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceProvidersRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "active_only",
            "activeOnly",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ActiveOnly,
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
                            "activeOnly" | "active_only" => Ok(GeneratedField::ActiveOnly),
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
            type Value = QueryComplianceProvidersRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceProvidersRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceProvidersRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut active_only__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ActiveOnly => {
                            if active_only__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activeOnly"));
                            }
                            active_only__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryComplianceProvidersRequest {
                    active_only: active_only__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceProvidersRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceProvidersResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceProvidersResponse", len)?;
        if !self.providers.is_empty() {
            struct_ser.serialize_field("providers", &self.providers)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceProvidersResponse {
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
            type Value = QueryComplianceProvidersResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceProvidersResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceProvidersResponse, V::Error>
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
                Ok(QueryComplianceProvidersResponse {
                    providers: providers__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceProvidersResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceStatusRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceStatusRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceStatusRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryComplianceStatusRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceStatusRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceStatusRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryComplianceStatusRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceStatusRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryComplianceStatusResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.record.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryComplianceStatusResponse", len)?;
        if let Some(v) = self.record.as_ref() {
            struct_ser.serialize_field("record", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryComplianceStatusResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "record",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Record,
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
                            "record" => Ok(GeneratedField::Record),
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
            type Value = QueryComplianceStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryComplianceStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryComplianceStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut record__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Record => {
                            if record__.is_some() {
                                return Err(serde::de::Error::duplicate_field("record"));
                            }
                            record__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryComplianceStatusResponse {
                    record: record__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryComplianceStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryConsentSettingsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.scope_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryConsentSettingsRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryConsentSettingsRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "scope_id",
            "scopeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ScopeId,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryConsentSettingsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryConsentSettingsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryConsentSettingsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut scope_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryConsentSettingsRequest {
                    account_address: account_address__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryConsentSettingsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryConsentSettingsResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.global_settings.is_some() {
            len += 1;
        }
        if !self.scope_consents.is_empty() {
            len += 1;
        }
        if self.consent_version != 0 {
            len += 1;
        }
        if self.last_updated_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryConsentSettingsResponse", len)?;
        if let Some(v) = self.global_settings.as_ref() {
            struct_ser.serialize_field("globalSettings", v)?;
        }
        if !self.scope_consents.is_empty() {
            struct_ser.serialize_field("scopeConsents", &self.scope_consents)?;
        }
        if self.consent_version != 0 {
            struct_ser.serialize_field("consentVersion", &self.consent_version)?;
        }
        if self.last_updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdatedAt", ToString::to_string(&self.last_updated_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryConsentSettingsResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "global_settings",
            "globalSettings",
            "scope_consents",
            "scopeConsents",
            "consent_version",
            "consentVersion",
            "last_updated_at",
            "lastUpdatedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            GlobalSettings,
            ScopeConsents,
            ConsentVersion,
            LastUpdatedAt,
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
                            "globalSettings" | "global_settings" => Ok(GeneratedField::GlobalSettings),
                            "scopeConsents" | "scope_consents" => Ok(GeneratedField::ScopeConsents),
                            "consentVersion" | "consent_version" => Ok(GeneratedField::ConsentVersion),
                            "lastUpdatedAt" | "last_updated_at" => Ok(GeneratedField::LastUpdatedAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryConsentSettingsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryConsentSettingsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryConsentSettingsResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut global_settings__ = None;
                let mut scope_consents__ = None;
                let mut consent_version__ = None;
                let mut last_updated_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::GlobalSettings => {
                            if global_settings__.is_some() {
                                return Err(serde::de::Error::duplicate_field("globalSettings"));
                            }
                            global_settings__ = map_.next_value()?;
                        }
                        GeneratedField::ScopeConsents => {
                            if scope_consents__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeConsents"));
                            }
                            scope_consents__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ConsentVersion => {
                            if consent_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentVersion"));
                            }
                            consent_version__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastUpdatedAt => {
                            if last_updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdatedAt"));
                            }
                            last_updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryConsentSettingsResponse {
                    global_settings: global_settings__,
                    scope_consents: scope_consents__.unwrap_or_default(),
                    consent_version: consent_version__.unwrap_or_default(),
                    last_updated_at: last_updated_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryConsentSettingsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDerivedFeatureHashesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.requester.is_empty() {
            len += 1;
        }
        if !self.purpose.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryDerivedFeatureHashesRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.requester.is_empty() {
            struct_ser.serialize_field("requester", &self.requester)?;
        }
        if !self.purpose.is_empty() {
            struct_ser.serialize_field("purpose", &self.purpose)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDerivedFeatureHashesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "requester",
            "purpose",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            Requester,
            Purpose,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "requester" => Ok(GeneratedField::Requester),
                            "purpose" => Ok(GeneratedField::Purpose),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDerivedFeatureHashesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryDerivedFeatureHashesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDerivedFeatureHashesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut requester__ = None;
                let mut purpose__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Requester => {
                            if requester__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requester"));
                            }
                            requester__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Purpose => {
                            if purpose__.is_some() {
                                return Err(serde::de::Error::duplicate_field("purpose"));
                            }
                            purpose__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDerivedFeatureHashesRequest {
                    account_address: account_address__.unwrap_or_default(),
                    requester: requester__.unwrap_or_default(),
                    purpose: purpose__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryDerivedFeatureHashesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDerivedFeatureHashesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.allowed {
            len += 1;
        }
        if !self.denial_reason.is_empty() {
            len += 1;
        }
        if !self.face_embedding_hash.is_empty() {
            len += 1;
        }
        if !self.doc_field_hashes.is_empty() {
            len += 1;
        }
        if !self.biometric_hash.is_empty() {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryDerivedFeatureHashesResponse", len)?;
        if self.allowed {
            struct_ser.serialize_field("allowed", &self.allowed)?;
        }
        if !self.denial_reason.is_empty() {
            struct_ser.serialize_field("denialReason", &self.denial_reason)?;
        }
        if !self.face_embedding_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("faceEmbeddingHash", pbjson::private::base64::encode(&self.face_embedding_hash).as_str())?;
        }
        if !self.doc_field_hashes.is_empty() {
            let v: std::collections::HashMap<_, _> = self.doc_field_hashes.iter()
                .map(|(k, v)| (k, pbjson::private::base64::encode(v))).collect();
            struct_ser.serialize_field("docFieldHashes", &v)?;
        }
        if !self.biometric_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("biometricHash", pbjson::private::base64::encode(&self.biometric_hash).as_str())?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDerivedFeatureHashesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "allowed",
            "denial_reason",
            "denialReason",
            "face_embedding_hash",
            "faceEmbeddingHash",
            "doc_field_hashes",
            "docFieldHashes",
            "biometric_hash",
            "biometricHash",
            "model_version",
            "modelVersion",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Allowed,
            DenialReason,
            FaceEmbeddingHash,
            DocFieldHashes,
            BiometricHash,
            ModelVersion,
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
                            "allowed" => Ok(GeneratedField::Allowed),
                            "denialReason" | "denial_reason" => Ok(GeneratedField::DenialReason),
                            "faceEmbeddingHash" | "face_embedding_hash" => Ok(GeneratedField::FaceEmbeddingHash),
                            "docFieldHashes" | "doc_field_hashes" => Ok(GeneratedField::DocFieldHashes),
                            "biometricHash" | "biometric_hash" => Ok(GeneratedField::BiometricHash),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDerivedFeatureHashesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryDerivedFeatureHashesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDerivedFeatureHashesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut allowed__ = None;
                let mut denial_reason__ = None;
                let mut face_embedding_hash__ = None;
                let mut doc_field_hashes__ = None;
                let mut biometric_hash__ = None;
                let mut model_version__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Allowed => {
                            if allowed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("allowed"));
                            }
                            allowed__ = Some(map_.next_value()?);
                        }
                        GeneratedField::DenialReason => {
                            if denial_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("denialReason"));
                            }
                            denial_reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::FaceEmbeddingHash => {
                            if face_embedding_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("faceEmbeddingHash"));
                            }
                            face_embedding_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DocFieldHashes => {
                            if doc_field_hashes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("docFieldHashes"));
                            }
                            doc_field_hashes__ = Some(
                                map_.next_value::<std::collections::HashMap<_, ::pbjson::private::BytesDeserialize<_>>>()?
                                    .into_iter().map(|(k,v)| (k, v.0)).collect()
                            );
                        }
                        GeneratedField::BiometricHash => {
                            if biometric_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("biometricHash"));
                            }
                            biometric_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDerivedFeatureHashesResponse {
                    allowed: allowed__.unwrap_or_default(),
                    denial_reason: denial_reason__.unwrap_or_default(),
                    face_embedding_hash: face_embedding_hash__.unwrap_or_default(),
                    doc_field_hashes: doc_field_hashes__.unwrap_or_default(),
                    biometric_hash: biometric_hash__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryDerivedFeatureHashesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDerivedFeaturesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryDerivedFeaturesRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDerivedFeaturesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryDerivedFeaturesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryDerivedFeaturesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDerivedFeaturesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDerivedFeaturesRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryDerivedFeaturesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryDerivedFeaturesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.features.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryDerivedFeaturesResponse", len)?;
        if let Some(v) = self.features.as_ref() {
            struct_ser.serialize_field("features", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryDerivedFeaturesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "features",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Features,
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
                            "features" => Ok(GeneratedField::Features),
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
            type Value = QueryDerivedFeaturesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryDerivedFeaturesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryDerivedFeaturesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut features__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Features => {
                            if features__.is_some() {
                                return Err(serde::de::Error::duplicate_field("features"));
                            }
                            features__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryDerivedFeaturesResponse {
                    features: features__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryDerivedFeaturesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityRecordRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityRecordRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityRecordRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryIdentityRecordRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityRecordRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityRecordRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityRecordRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityRecordRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityRecordResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.record.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityRecordResponse", len)?;
        if let Some(v) = self.record.as_ref() {
            struct_ser.serialize_field("record", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityRecordResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "record",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = QueryIdentityRecordResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityRecordResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityRecordResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut record__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Record => {
                            if record__.is_some() {
                                return Err(serde::de::Error::duplicate_field("record"));
                            }
                            record__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryIdentityRecordResponse {
                    record: record__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityRecordResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryIdentityRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.identity.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityResponse", len)?;
        if let Some(v) = self.identity.as_ref() {
            struct_ser.serialize_field("identity", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "identity",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Identity,
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
                            "identity" => Ok(GeneratedField::Identity),
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
            type Value = QueryIdentityResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut identity__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Identity => {
                            if identity__.is_some() {
                                return Err(serde::de::Error::duplicate_field("identity"));
                            }
                            identity__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityResponse {
                    identity: identity__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityScoreRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityScoreRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityScoreRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryIdentityScoreRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityScoreRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityScoreRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityScoreRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityScoreRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityScoreResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.score.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityScoreResponse", len)?;
        if let Some(v) = self.score.as_ref() {
            struct_ser.serialize_field("score", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityScoreResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "score",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Score,
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
                            "score" => Ok(GeneratedField::Score),
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
            type Value = QueryIdentityScoreResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityScoreResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityScoreResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut score__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityScoreResponse {
                    score: score__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityScoreResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityStatusRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityStatusRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityStatusRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryIdentityStatusRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityStatusRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityStatusRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityStatusRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityStatusRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityStatusResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.tier != 0 {
            len += 1;
        }
        if self.score != 0 {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        if self.last_updated_at != 0 {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityStatusResponse", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.status != 0 {
            let v = AccountStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.tier != 0 {
            let v = IdentityTier::try_from(self.tier)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.tier)))?;
            struct_ser.serialize_field("tier", &v)?;
        }
        if self.score != 0 {
            struct_ser.serialize_field("score", &self.score)?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        if self.last_updated_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastUpdatedAt", ToString::to_string(&self.last_updated_at).as_str())?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityStatusResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "status",
            "tier",
            "score",
            "model_version",
            "modelVersion",
            "last_updated_at",
            "lastUpdatedAt",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            Status,
            Tier,
            Score,
            ModelVersion,
            LastUpdatedAt,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "status" => Ok(GeneratedField::Status),
                            "tier" => Ok(GeneratedField::Tier),
                            "score" => Ok(GeneratedField::Score),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            "lastUpdatedAt" | "last_updated_at" => Ok(GeneratedField::LastUpdatedAt),
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
            type Value = QueryIdentityStatusResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityStatusResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityStatusResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut status__ = None;
                let mut tier__ = None;
                let mut score__ = None;
                let mut model_version__ = None;
                let mut last_updated_at__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::Tier => {
                            if tier__.is_some() {
                                return Err(serde::de::Error::duplicate_field("tier"));
                            }
                            tier__ = Some(map_.next_value::<IdentityTier>()? as i32);
                        }
                        GeneratedField::Score => {
                            if score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("score"));
                            }
                            score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LastUpdatedAt => {
                            if last_updated_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastUpdatedAt"));
                            }
                            last_updated_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityStatusResponse {
                    account_address: account_address__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    tier: tier__.unwrap_or_default(),
                    score: score__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                    last_updated_at: last_updated_at__.unwrap_or_default(),
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityStatusResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityWalletRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityWalletRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityWalletRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryIdentityWalletRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityWalletRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityWalletRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityWalletRequest {
                    account_address: account_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityWalletRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryIdentityWalletResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.wallet.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryIdentityWalletResponse", len)?;
        if let Some(v) = self.wallet.as_ref() {
            struct_ser.serialize_field("wallet", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryIdentityWalletResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "wallet",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Wallet,
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
                            "wallet" => Ok(GeneratedField::Wallet),
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
            type Value = QueryIdentityWalletResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryIdentityWalletResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryIdentityWalletResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut wallet__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Wallet => {
                            if wallet__.is_some() {
                                return Err(serde::de::Error::duplicate_field("wallet"));
                            }
                            wallet__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryIdentityWalletResponse {
                    wallet: wallet__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryIdentityWalletResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModelHistoryRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.model_type.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryModelHistoryRequest", len)?;
        if !self.model_type.is_empty() {
            struct_ser.serialize_field("modelType", &self.model_type)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModelHistoryRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "model_type",
            "modelType",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ModelType,
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
                            "modelType" | "model_type" => Ok(GeneratedField::ModelType),
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
            type Value = QueryModelHistoryRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryModelHistoryRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModelHistoryRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut model_type__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ModelType => {
                            if model_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelType"));
                            }
                            model_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryModelHistoryRequest {
                    model_type: model_type__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryModelHistoryRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModelHistoryResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.history.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryModelHistoryResponse", len)?;
        if !self.history.is_empty() {
            struct_ser.serialize_field("history", &self.history)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModelHistoryResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "history",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            History,
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
                            "history" => Ok(GeneratedField::History),
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
            type Value = QueryModelHistoryResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryModelHistoryResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModelHistoryResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut history__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::History => {
                            if history__.is_some() {
                                return Err(serde::de::Error::duplicate_field("history"));
                            }
                            history__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryModelHistoryResponse {
                    history: history__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryModelHistoryResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModelParamsRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryModelParamsRequest", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModelParamsRequest {
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
            type Value = QueryModelParamsRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryModelParamsRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModelParamsRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(QueryModelParamsRequest {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryModelParamsRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModelParamsResponse {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryModelParamsResponse", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModelParamsResponse {
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
            type Value = QueryModelParamsResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryModelParamsResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModelParamsResponse, V::Error>
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
                Ok(QueryModelParamsResponse {
                    params: params__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryModelParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModelVersionRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.model_type.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryModelVersionRequest", len)?;
        if !self.model_type.is_empty() {
            struct_ser.serialize_field("modelType", &self.model_type)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModelVersionRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "model_type",
            "modelType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ModelType,
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
                            "modelType" | "model_type" => Ok(GeneratedField::ModelType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryModelVersionRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryModelVersionRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModelVersionRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut model_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ModelType => {
                            if model_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelType"));
                            }
                            model_type__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryModelVersionRequest {
                    model_type: model_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryModelVersionRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryModelVersionResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.model_info.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryModelVersionResponse", len)?;
        if let Some(v) = self.model_info.as_ref() {
            struct_ser.serialize_field("modelInfo", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryModelVersionResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "model_info",
            "modelInfo",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ModelInfo,
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
                            "modelInfo" | "model_info" => Ok(GeneratedField::ModelInfo),
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
            type Value = QueryModelVersionResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryModelVersionResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryModelVersionResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut model_info__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ModelInfo => {
                            if model_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelInfo"));
                            }
                            model_info__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryModelVersionResponse {
                    model_info: model_info__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryModelVersionResponse", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryParamsRequest", len)?;
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
                formatter.write_str("struct virtengine.veid.v1.QueryParamsRequest")
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
        deserializer.deserialize_struct("virtengine.veid.v1.QueryParamsRequest", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.veid.v1.QueryParamsResponse")
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
        deserializer.deserialize_struct("virtengine.veid.v1.QueryParamsResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryScopeRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if !self.scope_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryScopeRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryScopeRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "scope_id",
            "scopeId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ScopeId,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryScopeRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryScopeRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryScopeRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut scope_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryScopeRequest {
                    account_address: account_address__.unwrap_or_default(),
                    scope_id: scope_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryScopeRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryScopeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.scope.is_some() {
            len += 1;
        }
        if self.found {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryScopeResponse", len)?;
        if let Some(v) = self.scope.as_ref() {
            struct_ser.serialize_field("scope", v)?;
        }
        if self.found {
            struct_ser.serialize_field("found", &self.found)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryScopeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope",
            "found",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Scope,
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
                            "scope" => Ok(GeneratedField::Scope),
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
            type Value = QueryScopeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryScopeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryScopeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope__ = None;
                let mut found__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Scope => {
                            if scope__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scope"));
                            }
                            scope__ = map_.next_value()?;
                        }
                        GeneratedField::Found => {
                            if found__.is_some() {
                                return Err(serde::de::Error::duplicate_field("found"));
                            }
                            found__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryScopeResponse {
                    scope: scope__,
                    found: found__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryScopeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryScopesByTypeRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryScopesByTypeRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryScopesByTypeRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "scope_type",
            "scopeType",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ScopeType,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryScopesByTypeRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryScopesByTypeRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryScopesByTypeRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut scope_type__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                    }
                }
                Ok(QueryScopesByTypeRequest {
                    account_address: account_address__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryScopesByTypeRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryScopesByTypeResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scopes.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryScopesByTypeResponse", len)?;
        if !self.scopes.is_empty() {
            struct_ser.serialize_field("scopes", &self.scopes)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryScopesByTypeResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scopes",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
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
            type Value = QueryScopesByTypeResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryScopesByTypeResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryScopesByTypeResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scopes__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Scopes => {
                            if scopes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopes"));
                            }
                            scopes__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryScopesByTypeResponse {
                    scopes: scopes__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryScopesByTypeResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryScopesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if self.status_filter != 0 {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryScopesRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if self.status_filter != 0 {
            let v = VerificationStatus::try_from(self.status_filter)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status_filter)))?;
            struct_ser.serialize_field("statusFilter", &v)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryScopesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "scope_type",
            "scopeType",
            "status_filter",
            "statusFilter",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ScopeType,
            StatusFilter,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "statusFilter" | "status_filter" => Ok(GeneratedField::StatusFilter),
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
            type Value = QueryScopesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryScopesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryScopesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut scope_type__ = None;
                let mut status_filter__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::StatusFilter => {
                            if status_filter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("statusFilter"));
                            }
                            status_filter__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryScopesRequest {
                    account_address: account_address__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    status_filter: status_filter__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryScopesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryScopesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scopes.is_empty() {
            len += 1;
        }
        if self.pagination.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryScopesResponse", len)?;
        if !self.scopes.is_empty() {
            struct_ser.serialize_field("scopes", &self.scopes)?;
        }
        if let Some(v) = self.pagination.as_ref() {
            struct_ser.serialize_field("pagination", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryScopesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scopes",
            "pagination",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Scopes,
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
                            "scopes" => Ok(GeneratedField::Scopes),
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
            type Value = QueryScopesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryScopesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryScopesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scopes__ = None;
                let mut pagination__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Scopes => {
                            if scopes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopes"));
                            }
                            scopes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Pagination => {
                            if pagination__.is_some() {
                                return Err(serde::de::Error::duplicate_field("pagination"));
                            }
                            pagination__ = map_.next_value()?;
                        }
                    }
                }
                Ok(QueryScopesResponse {
                    scopes: scopes__.unwrap_or_default(),
                    pagination: pagination__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryScopesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidatorModelSyncRequest {
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
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryValidatorModelSyncRequest", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidatorModelSyncRequest {
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
            type Value = QueryValidatorModelSyncRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryValidatorModelSyncRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidatorModelSyncRequest, V::Error>
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
                Ok(QueryValidatorModelSyncRequest {
                    validator_address: validator_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryValidatorModelSyncRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryValidatorModelSyncResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.report.is_some() {
            len += 1;
        }
        if self.is_synced {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryValidatorModelSyncResponse", len)?;
        if let Some(v) = self.report.as_ref() {
            struct_ser.serialize_field("report", v)?;
        }
        if self.is_synced {
            struct_ser.serialize_field("isSynced", &self.is_synced)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryValidatorModelSyncResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "report",
            "is_synced",
            "isSynced",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Report,
            IsSynced,
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
                            "report" => Ok(GeneratedField::Report),
                            "isSynced" | "is_synced" => Ok(GeneratedField::IsSynced),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryValidatorModelSyncResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryValidatorModelSyncResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryValidatorModelSyncResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut report__ = None;
                let mut is_synced__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Report => {
                            if report__.is_some() {
                                return Err(serde::de::Error::duplicate_field("report"));
                            }
                            report__ = map_.next_value()?;
                        }
                        GeneratedField::IsSynced => {
                            if is_synced__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isSynced"));
                            }
                            is_synced__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryValidatorModelSyncResponse {
                    report: report__,
                    is_synced: is_synced__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryValidatorModelSyncResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryVerificationHistoryRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.limit != 0 {
            len += 1;
        }
        if self.offset != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryVerificationHistoryRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.limit != 0 {
            struct_ser.serialize_field("limit", &self.limit)?;
        }
        if self.offset != 0 {
            struct_ser.serialize_field("offset", &self.offset)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryVerificationHistoryRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "limit",
            "offset",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            Limit,
            Offset,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "limit" => Ok(GeneratedField::Limit),
                            "offset" => Ok(GeneratedField::Offset),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryVerificationHistoryRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryVerificationHistoryRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryVerificationHistoryRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut limit__ = None;
                let mut offset__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Limit => {
                            if limit__.is_some() {
                                return Err(serde::de::Error::duplicate_field("limit"));
                            }
                            limit__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Offset => {
                            if offset__.is_some() {
                                return Err(serde::de::Error::duplicate_field("offset"));
                            }
                            offset__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryVerificationHistoryRequest {
                    account_address: account_address__.unwrap_or_default(),
                    limit: limit__.unwrap_or_default(),
                    offset: offset__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryVerificationHistoryRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryVerificationHistoryResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.entries.is_empty() {
            len += 1;
        }
        if self.total_count != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryVerificationHistoryResponse", len)?;
        if !self.entries.is_empty() {
            struct_ser.serialize_field("entries", &self.entries)?;
        }
        if self.total_count != 0 {
            struct_ser.serialize_field("totalCount", &self.total_count)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryVerificationHistoryResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "entries",
            "total_count",
            "totalCount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Entries,
            TotalCount,
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
                            "entries" => Ok(GeneratedField::Entries),
                            "totalCount" | "total_count" => Ok(GeneratedField::TotalCount),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryVerificationHistoryResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryVerificationHistoryResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryVerificationHistoryResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut entries__ = None;
                let mut total_count__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Entries => {
                            if entries__.is_some() {
                                return Err(serde::de::Error::duplicate_field("entries"));
                            }
                            entries__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalCount => {
                            if total_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalCount"));
                            }
                            total_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryVerificationHistoryResponse {
                    entries: entries__.unwrap_or_default(),
                    total_count: total_count__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryVerificationHistoryResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryWalletScopesRequest {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.account_address.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if self.status_filter != 0 {
            len += 1;
        }
        if self.active_only {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryWalletScopesRequest", len)?;
        if !self.account_address.is_empty() {
            struct_ser.serialize_field("accountAddress", &self.account_address)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if self.status_filter != 0 {
            let v = VerificationStatus::try_from(self.status_filter)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status_filter)))?;
            struct_ser.serialize_field("statusFilter", &v)?;
        }
        if self.active_only {
            struct_ser.serialize_field("activeOnly", &self.active_only)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryWalletScopesRequest {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "account_address",
            "accountAddress",
            "scope_type",
            "scopeType",
            "status_filter",
            "statusFilter",
            "active_only",
            "activeOnly",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            AccountAddress,
            ScopeType,
            StatusFilter,
            ActiveOnly,
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
                            "accountAddress" | "account_address" => Ok(GeneratedField::AccountAddress),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "statusFilter" | "status_filter" => Ok(GeneratedField::StatusFilter),
                            "activeOnly" | "active_only" => Ok(GeneratedField::ActiveOnly),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryWalletScopesRequest;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryWalletScopesRequest")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryWalletScopesRequest, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut account_address__ = None;
                let mut scope_type__ = None;
                let mut status_filter__ = None;
                let mut active_only__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::AccountAddress => {
                            if account_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("accountAddress"));
                            }
                            account_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::StatusFilter => {
                            if status_filter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("statusFilter"));
                            }
                            status_filter__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::ActiveOnly => {
                            if active_only__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activeOnly"));
                            }
                            active_only__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(QueryWalletScopesRequest {
                    account_address: account_address__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    status_filter: status_filter__.unwrap_or_default(),
                    active_only: active_only__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryWalletScopesRequest", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueryWalletScopesResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scopes.is_empty() {
            len += 1;
        }
        if self.total_count != 0 {
            len += 1;
        }
        if self.active_count != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.QueryWalletScopesResponse", len)?;
        if !self.scopes.is_empty() {
            struct_ser.serialize_field("scopes", &self.scopes)?;
        }
        if self.total_count != 0 {
            struct_ser.serialize_field("totalCount", &self.total_count)?;
        }
        if self.active_count != 0 {
            struct_ser.serialize_field("activeCount", &self.active_count)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueryWalletScopesResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scopes",
            "total_count",
            "totalCount",
            "active_count",
            "activeCount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Scopes,
            TotalCount,
            ActiveCount,
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
                            "scopes" => Ok(GeneratedField::Scopes),
                            "totalCount" | "total_count" => Ok(GeneratedField::TotalCount),
                            "activeCount" | "active_count" => Ok(GeneratedField::ActiveCount),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueryWalletScopesResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.QueryWalletScopesResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<QueryWalletScopesResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scopes__ = None;
                let mut total_count__ = None;
                let mut active_count__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Scopes => {
                            if scopes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopes"));
                            }
                            scopes__ = Some(map_.next_value()?);
                        }
                        GeneratedField::TotalCount => {
                            if total_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("totalCount"));
                            }
                            total_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ActiveCount => {
                            if active_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("activeCount"));
                            }
                            active_count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(QueryWalletScopesResponse {
                    scopes: scopes__.unwrap_or_default(),
                    total_count: total_count__.unwrap_or_default(),
                    active_count: active_count__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.QueryWalletScopesResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ScopeConsent {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.granted {
            len += 1;
        }
        if self.granted_at != 0 {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if !self.purpose.is_empty() {
            len += 1;
        }
        if !self.granted_to_providers.is_empty() {
            len += 1;
        }
        if !self.restrictions.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ScopeConsent", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.granted {
            struct_ser.serialize_field("granted", &self.granted)?;
        }
        if self.granted_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("grantedAt", ToString::to_string(&self.granted_at).as_str())?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if !self.purpose.is_empty() {
            struct_ser.serialize_field("purpose", &self.purpose)?;
        }
        if !self.granted_to_providers.is_empty() {
            struct_ser.serialize_field("grantedToProviders", &self.granted_to_providers)?;
        }
        if !self.restrictions.is_empty() {
            struct_ser.serialize_field("restrictions", &self.restrictions)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ScopeConsent {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "granted",
            "granted_at",
            "grantedAt",
            "revoked_at",
            "revokedAt",
            "expires_at",
            "expiresAt",
            "purpose",
            "granted_to_providers",
            "grantedToProviders",
            "restrictions",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            Granted,
            GrantedAt,
            RevokedAt,
            ExpiresAt,
            Purpose,
            GrantedToProviders,
            Restrictions,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "granted" => Ok(GeneratedField::Granted),
                            "grantedAt" | "granted_at" => Ok(GeneratedField::GrantedAt),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "purpose" => Ok(GeneratedField::Purpose),
                            "grantedToProviders" | "granted_to_providers" => Ok(GeneratedField::GrantedToProviders),
                            "restrictions" => Ok(GeneratedField::Restrictions),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ScopeConsent;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ScopeConsent")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ScopeConsent, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut granted__ = None;
                let mut granted_at__ = None;
                let mut revoked_at__ = None;
                let mut expires_at__ = None;
                let mut purpose__ = None;
                let mut granted_to_providers__ = None;
                let mut restrictions__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Granted => {
                            if granted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("granted"));
                            }
                            granted__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GrantedAt => {
                            if granted_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("grantedAt"));
                            }
                            granted_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Purpose => {
                            if purpose__.is_some() {
                                return Err(serde::de::Error::duplicate_field("purpose"));
                            }
                            purpose__ = Some(map_.next_value()?);
                        }
                        GeneratedField::GrantedToProviders => {
                            if granted_to_providers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("grantedToProviders"));
                            }
                            granted_to_providers__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Restrictions => {
                            if restrictions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("restrictions"));
                            }
                            restrictions__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ScopeConsent {
                    scope_id: scope_id__.unwrap_or_default(),
                    granted: granted__.unwrap_or_default(),
                    granted_at: granted_at__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    purpose: purpose__.unwrap_or_default(),
                    granted_to_providers: granted_to_providers__.unwrap_or_default(),
                    restrictions: restrictions__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ScopeConsent", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ScopeRef {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.uploaded_at != 0 {
            len += 1;
        }
        if self.verified_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ScopeRef", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if self.status != 0 {
            let v = VerificationStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.uploaded_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("uploadedAt", ToString::to_string(&self.uploaded_at).as_str())?;
        }
        if self.verified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("verifiedAt", ToString::to_string(&self.verified_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ScopeRef {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "scope_type",
            "scopeType",
            "status",
            "uploaded_at",
            "uploadedAt",
            "verified_at",
            "verifiedAt",
            "expires_at",
            "expiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            ScopeType,
            Status,
            UploadedAt,
            VerifiedAt,
            ExpiresAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "status" => Ok(GeneratedField::Status),
                            "uploadedAt" | "uploaded_at" => Ok(GeneratedField::UploadedAt),
                            "verifiedAt" | "verified_at" => Ok(GeneratedField::VerifiedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ScopeRef;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ScopeRef")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ScopeRef, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut scope_type__ = None;
                let mut status__ = None;
                let mut uploaded_at__ = None;
                let mut verified_at__ = None;
                let mut expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::UploadedAt => {
                            if uploaded_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uploadedAt"));
                            }
                            uploaded_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VerifiedAt => {
                            if verified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedAt"));
                            }
                            verified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ScopeRef {
                    scope_id: scope_id__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    uploaded_at: uploaded_at__.unwrap_or_default(),
                    verified_at: verified_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ScopeRef", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ScopeRefStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "SCOPE_REF_STATUS_UNSPECIFIED",
            Self::Active => "SCOPE_REF_STATUS_ACTIVE",
            Self::Revoked => "SCOPE_REF_STATUS_REVOKED",
            Self::Expired => "SCOPE_REF_STATUS_EXPIRED",
            Self::Pending => "SCOPE_REF_STATUS_PENDING",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ScopeRefStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "SCOPE_REF_STATUS_UNSPECIFIED",
            "SCOPE_REF_STATUS_ACTIVE",
            "SCOPE_REF_STATUS_REVOKED",
            "SCOPE_REF_STATUS_EXPIRED",
            "SCOPE_REF_STATUS_PENDING",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ScopeRefStatus;

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
                    "SCOPE_REF_STATUS_UNSPECIFIED" => Ok(ScopeRefStatus::Unspecified),
                    "SCOPE_REF_STATUS_ACTIVE" => Ok(ScopeRefStatus::Active),
                    "SCOPE_REF_STATUS_REVOKED" => Ok(ScopeRefStatus::Revoked),
                    "SCOPE_REF_STATUS_EXPIRED" => Ok(ScopeRefStatus::Expired),
                    "SCOPE_REF_STATUS_PENDING" => Ok(ScopeRefStatus::Pending),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for ScopeReference {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if !self.envelope_hash.is_empty() {
            len += 1;
        }
        if self.added_at != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.consent_granted {
            len += 1;
        }
        if !self.revocation_reason.is_empty() {
            len += 1;
        }
        if self.revoked_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ScopeReference", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if !self.envelope_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("envelopeHash", pbjson::private::base64::encode(&self.envelope_hash).as_str())?;
        }
        if self.added_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("addedAt", ToString::to_string(&self.added_at).as_str())?;
        }
        if self.status != 0 {
            let v = ScopeRefStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.consent_granted {
            struct_ser.serialize_field("consentGranted", &self.consent_granted)?;
        }
        if !self.revocation_reason.is_empty() {
            struct_ser.serialize_field("revocationReason", &self.revocation_reason)?;
        }
        if self.revoked_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("revokedAt", ToString::to_string(&self.revoked_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ScopeReference {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "scope_type",
            "scopeType",
            "envelope_hash",
            "envelopeHash",
            "added_at",
            "addedAt",
            "status",
            "consent_granted",
            "consentGranted",
            "revocation_reason",
            "revocationReason",
            "revoked_at",
            "revokedAt",
            "expires_at",
            "expiresAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            ScopeType,
            EnvelopeHash,
            AddedAt,
            Status,
            ConsentGranted,
            RevocationReason,
            RevokedAt,
            ExpiresAt,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "envelopeHash" | "envelope_hash" => Ok(GeneratedField::EnvelopeHash),
                            "addedAt" | "added_at" => Ok(GeneratedField::AddedAt),
                            "status" => Ok(GeneratedField::Status),
                            "consentGranted" | "consent_granted" => Ok(GeneratedField::ConsentGranted),
                            "revocationReason" | "revocation_reason" => Ok(GeneratedField::RevocationReason),
                            "revokedAt" | "revoked_at" => Ok(GeneratedField::RevokedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ScopeReference;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ScopeReference")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ScopeReference, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut scope_type__ = None;
                let mut envelope_hash__ = None;
                let mut added_at__ = None;
                let mut status__ = None;
                let mut consent_granted__ = None;
                let mut revocation_reason__ = None;
                let mut revoked_at__ = None;
                let mut expires_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::EnvelopeHash => {
                            if envelope_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("envelopeHash"));
                            }
                            envelope_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::AddedAt => {
                            if added_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addedAt"));
                            }
                            added_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<ScopeRefStatus>()? as i32);
                        }
                        GeneratedField::ConsentGranted => {
                            if consent_granted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentGranted"));
                            }
                            consent_granted__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevocationReason => {
                            if revocation_reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revocationReason"));
                            }
                            revocation_reason__ = Some(map_.next_value()?);
                        }
                        GeneratedField::RevokedAt => {
                            if revoked_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("revokedAt"));
                            }
                            revoked_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ScopeReference {
                    scope_id: scope_id__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    envelope_hash: envelope_hash__.unwrap_or_default(),
                    added_at: added_at__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    consent_granted: consent_granted__.unwrap_or_default(),
                    revocation_reason: revocation_reason__.unwrap_or_default(),
                    revoked_at: revoked_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ScopeReference", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ScopeType {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "SCOPE_TYPE_UNSPECIFIED",
            Self::IdDocument => "SCOPE_TYPE_ID_DOCUMENT",
            Self::Selfie => "SCOPE_TYPE_SELFIE",
            Self::FaceVideo => "SCOPE_TYPE_FACE_VIDEO",
            Self::Biometric => "SCOPE_TYPE_BIOMETRIC",
            Self::SsoMetadata => "SCOPE_TYPE_SSO_METADATA",
            Self::EmailProof => "SCOPE_TYPE_EMAIL_PROOF",
            Self::SmsProof => "SCOPE_TYPE_SMS_PROOF",
            Self::DomainVerify => "SCOPE_TYPE_DOMAIN_VERIFY",
            Self::AdSso => "SCOPE_TYPE_AD_SSO",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for ScopeType {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "SCOPE_TYPE_UNSPECIFIED",
            "SCOPE_TYPE_ID_DOCUMENT",
            "SCOPE_TYPE_SELFIE",
            "SCOPE_TYPE_FACE_VIDEO",
            "SCOPE_TYPE_BIOMETRIC",
            "SCOPE_TYPE_SSO_METADATA",
            "SCOPE_TYPE_EMAIL_PROOF",
            "SCOPE_TYPE_SMS_PROOF",
            "SCOPE_TYPE_DOMAIN_VERIFY",
            "SCOPE_TYPE_AD_SSO",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ScopeType;

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
                    "SCOPE_TYPE_UNSPECIFIED" => Ok(ScopeType::Unspecified),
                    "SCOPE_TYPE_ID_DOCUMENT" => Ok(ScopeType::IdDocument),
                    "SCOPE_TYPE_SELFIE" => Ok(ScopeType::Selfie),
                    "SCOPE_TYPE_FACE_VIDEO" => Ok(ScopeType::FaceVideo),
                    "SCOPE_TYPE_BIOMETRIC" => Ok(ScopeType::Biometric),
                    "SCOPE_TYPE_SSO_METADATA" => Ok(ScopeType::SsoMetadata),
                    "SCOPE_TYPE_EMAIL_PROOF" => Ok(ScopeType::EmailProof),
                    "SCOPE_TYPE_SMS_PROOF" => Ok(ScopeType::SmsProof),
                    "SCOPE_TYPE_DOMAIN_VERIFY" => Ok(ScopeType::DomainVerify),
                    "SCOPE_TYPE_AD_SSO" => Ok(ScopeType::AdSso),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for UploadMetadata {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.salt.is_empty() {
            len += 1;
        }
        if !self.salt_hash.is_empty() {
            len += 1;
        }
        if !self.device_fingerprint.is_empty() {
            len += 1;
        }
        if !self.client_id.is_empty() {
            len += 1;
        }
        if !self.client_signature.is_empty() {
            len += 1;
        }
        if !self.user_signature.is_empty() {
            len += 1;
        }
        if !self.payload_hash.is_empty() {
            len += 1;
        }
        if !self.upload_nonce.is_empty() {
            len += 1;
        }
        if self.capture_timestamp != 0 {
            len += 1;
        }
        if !self.geo_hint.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.UploadMetadata", len)?;
        if !self.salt.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("salt", pbjson::private::base64::encode(&self.salt).as_str())?;
        }
        if !self.salt_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("saltHash", pbjson::private::base64::encode(&self.salt_hash).as_str())?;
        }
        if !self.device_fingerprint.is_empty() {
            struct_ser.serialize_field("deviceFingerprint", &self.device_fingerprint)?;
        }
        if !self.client_id.is_empty() {
            struct_ser.serialize_field("clientId", &self.client_id)?;
        }
        if !self.client_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("clientSignature", pbjson::private::base64::encode(&self.client_signature).as_str())?;
        }
        if !self.user_signature.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("userSignature", pbjson::private::base64::encode(&self.user_signature).as_str())?;
        }
        if !self.payload_hash.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("payloadHash", pbjson::private::base64::encode(&self.payload_hash).as_str())?;
        }
        if !self.upload_nonce.is_empty() {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("uploadNonce", pbjson::private::base64::encode(&self.upload_nonce).as_str())?;
        }
        if self.capture_timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("captureTimestamp", ToString::to_string(&self.capture_timestamp).as_str())?;
        }
        if !self.geo_hint.is_empty() {
            struct_ser.serialize_field("geoHint", &self.geo_hint)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for UploadMetadata {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "salt",
            "salt_hash",
            "saltHash",
            "device_fingerprint",
            "deviceFingerprint",
            "client_id",
            "clientId",
            "client_signature",
            "clientSignature",
            "user_signature",
            "userSignature",
            "payload_hash",
            "payloadHash",
            "upload_nonce",
            "uploadNonce",
            "capture_timestamp",
            "captureTimestamp",
            "geo_hint",
            "geoHint",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Salt,
            SaltHash,
            DeviceFingerprint,
            ClientId,
            ClientSignature,
            UserSignature,
            PayloadHash,
            UploadNonce,
            CaptureTimestamp,
            GeoHint,
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
                            "salt" => Ok(GeneratedField::Salt),
                            "saltHash" | "salt_hash" => Ok(GeneratedField::SaltHash),
                            "deviceFingerprint" | "device_fingerprint" => Ok(GeneratedField::DeviceFingerprint),
                            "clientId" | "client_id" => Ok(GeneratedField::ClientId),
                            "clientSignature" | "client_signature" => Ok(GeneratedField::ClientSignature),
                            "userSignature" | "user_signature" => Ok(GeneratedField::UserSignature),
                            "payloadHash" | "payload_hash" => Ok(GeneratedField::PayloadHash),
                            "uploadNonce" | "upload_nonce" => Ok(GeneratedField::UploadNonce),
                            "captureTimestamp" | "capture_timestamp" => Ok(GeneratedField::CaptureTimestamp),
                            "geoHint" | "geo_hint" => Ok(GeneratedField::GeoHint),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = UploadMetadata;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.UploadMetadata")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<UploadMetadata, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut salt__ = None;
                let mut salt_hash__ = None;
                let mut device_fingerprint__ = None;
                let mut client_id__ = None;
                let mut client_signature__ = None;
                let mut user_signature__ = None;
                let mut payload_hash__ = None;
                let mut upload_nonce__ = None;
                let mut capture_timestamp__ = None;
                let mut geo_hint__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Salt => {
                            if salt__.is_some() {
                                return Err(serde::de::Error::duplicate_field("salt"));
                            }
                            salt__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SaltHash => {
                            if salt_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("saltHash"));
                            }
                            salt_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::DeviceFingerprint => {
                            if device_fingerprint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("deviceFingerprint"));
                            }
                            device_fingerprint__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClientId => {
                            if client_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientId"));
                            }
                            client_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ClientSignature => {
                            if client_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("clientSignature"));
                            }
                            client_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UserSignature => {
                            if user_signature__.is_some() {
                                return Err(serde::de::Error::duplicate_field("userSignature"));
                            }
                            user_signature__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PayloadHash => {
                            if payload_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("payloadHash"));
                            }
                            payload_hash__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::UploadNonce => {
                            if upload_nonce__.is_some() {
                                return Err(serde::de::Error::duplicate_field("uploadNonce"));
                            }
                            upload_nonce__ = 
                                Some(map_.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::CaptureTimestamp => {
                            if capture_timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("captureTimestamp"));
                            }
                            capture_timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::GeoHint => {
                            if geo_hint__.is_some() {
                                return Err(serde::de::Error::duplicate_field("geoHint"));
                            }
                            geo_hint__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(UploadMetadata {
                    salt: salt__.unwrap_or_default(),
                    salt_hash: salt_hash__.unwrap_or_default(),
                    device_fingerprint: device_fingerprint__.unwrap_or_default(),
                    client_id: client_id__.unwrap_or_default(),
                    client_signature: client_signature__.unwrap_or_default(),
                    user_signature: user_signature__.unwrap_or_default(),
                    payload_hash: payload_hash__.unwrap_or_default(),
                    upload_nonce: upload_nonce__.unwrap_or_default(),
                    capture_timestamp: capture_timestamp__.unwrap_or_default(),
                    geo_hint: geo_hint__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.UploadMetadata", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorModelReport {
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
        if !self.model_versions.is_empty() {
            len += 1;
        }
        if self.reported_at != 0 {
            len += 1;
        }
        if self.last_verified != 0 {
            len += 1;
        }
        if self.is_synced {
            len += 1;
        }
        if !self.mismatched_models.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.ValidatorModelReport", len)?;
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.model_versions.is_empty() {
            struct_ser.serialize_field("modelVersions", &self.model_versions)?;
        }
        if self.reported_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reportedAt", ToString::to_string(&self.reported_at).as_str())?;
        }
        if self.last_verified != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("lastVerified", ToString::to_string(&self.last_verified).as_str())?;
        }
        if self.is_synced {
            struct_ser.serialize_field("isSynced", &self.is_synced)?;
        }
        if !self.mismatched_models.is_empty() {
            struct_ser.serialize_field("mismatchedModels", &self.mismatched_models)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValidatorModelReport {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_address",
            "validatorAddress",
            "model_versions",
            "modelVersions",
            "reported_at",
            "reportedAt",
            "last_verified",
            "lastVerified",
            "is_synced",
            "isSynced",
            "mismatched_models",
            "mismatchedModels",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorAddress,
            ModelVersions,
            ReportedAt,
            LastVerified,
            IsSynced,
            MismatchedModels,
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
                            "modelVersions" | "model_versions" => Ok(GeneratedField::ModelVersions),
                            "reportedAt" | "reported_at" => Ok(GeneratedField::ReportedAt),
                            "lastVerified" | "last_verified" => Ok(GeneratedField::LastVerified),
                            "isSynced" | "is_synced" => Ok(GeneratedField::IsSynced),
                            "mismatchedModels" | "mismatched_models" => Ok(GeneratedField::MismatchedModels),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValidatorModelReport;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.ValidatorModelReport")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ValidatorModelReport, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_address__ = None;
                let mut model_versions__ = None;
                let mut reported_at__ = None;
                let mut last_verified__ = None;
                let mut is_synced__ = None;
                let mut mismatched_models__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ValidatorAddress => {
                            if validator_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorAddress"));
                            }
                            validator_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelVersions => {
                            if model_versions__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersions"));
                            }
                            model_versions__ = Some(
                                map_.next_value::<std::collections::HashMap<_, _>>()?
                            );
                        }
                        GeneratedField::ReportedAt => {
                            if reported_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reportedAt"));
                            }
                            reported_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastVerified => {
                            if last_verified__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastVerified"));
                            }
                            last_verified__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::IsSynced => {
                            if is_synced__.is_some() {
                                return Err(serde::de::Error::duplicate_field("isSynced"));
                            }
                            is_synced__ = Some(map_.next_value()?);
                        }
                        GeneratedField::MismatchedModels => {
                            if mismatched_models__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mismatchedModels"));
                            }
                            mismatched_models__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ValidatorModelReport {
                    validator_address: validator_address__.unwrap_or_default(),
                    model_versions: model_versions__.unwrap_or_default(),
                    reported_at: reported_at__.unwrap_or_default(),
                    last_verified: last_verified__.unwrap_or_default(),
                    is_synced: is_synced__.unwrap_or_default(),
                    mismatched_models: mismatched_models__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.ValidatorModelReport", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for VerificationHistoryEntry {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.entry_id.is_empty() {
            len += 1;
        }
        if self.timestamp != 0 {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.previous_score != 0 {
            len += 1;
        }
        if self.new_score != 0 {
            len += 1;
        }
        if self.previous_status != 0 {
            len += 1;
        }
        if self.new_status != 0 {
            len += 1;
        }
        if !self.scopes_evaluated.is_empty() {
            len += 1;
        }
        if !self.model_version.is_empty() {
            len += 1;
        }
        if !self.validator_address.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.VerificationHistoryEntry", len)?;
        if !self.entry_id.is_empty() {
            struct_ser.serialize_field("entryId", &self.entry_id)?;
        }
        if self.timestamp != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("timestamp", ToString::to_string(&self.timestamp).as_str())?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if self.previous_score != 0 {
            struct_ser.serialize_field("previousScore", &self.previous_score)?;
        }
        if self.new_score != 0 {
            struct_ser.serialize_field("newScore", &self.new_score)?;
        }
        if self.previous_status != 0 {
            let v = AccountStatus::try_from(self.previous_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.previous_status)))?;
            struct_ser.serialize_field("previousStatus", &v)?;
        }
        if self.new_status != 0 {
            let v = AccountStatus::try_from(self.new_status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.new_status)))?;
            struct_ser.serialize_field("newStatus", &v)?;
        }
        if !self.scopes_evaluated.is_empty() {
            struct_ser.serialize_field("scopesEvaluated", &self.scopes_evaluated)?;
        }
        if !self.model_version.is_empty() {
            struct_ser.serialize_field("modelVersion", &self.model_version)?;
        }
        if !self.validator_address.is_empty() {
            struct_ser.serialize_field("validatorAddress", &self.validator_address)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for VerificationHistoryEntry {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "entry_id",
            "entryId",
            "timestamp",
            "block_height",
            "blockHeight",
            "previous_score",
            "previousScore",
            "new_score",
            "newScore",
            "previous_status",
            "previousStatus",
            "new_status",
            "newStatus",
            "scopes_evaluated",
            "scopesEvaluated",
            "model_version",
            "modelVersion",
            "validator_address",
            "validatorAddress",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EntryId,
            Timestamp,
            BlockHeight,
            PreviousScore,
            NewScore,
            PreviousStatus,
            NewStatus,
            ScopesEvaluated,
            ModelVersion,
            ValidatorAddress,
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
                            "entryId" | "entry_id" => Ok(GeneratedField::EntryId),
                            "timestamp" => Ok(GeneratedField::Timestamp),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "previousScore" | "previous_score" => Ok(GeneratedField::PreviousScore),
                            "newScore" | "new_score" => Ok(GeneratedField::NewScore),
                            "previousStatus" | "previous_status" => Ok(GeneratedField::PreviousStatus),
                            "newStatus" | "new_status" => Ok(GeneratedField::NewStatus),
                            "scopesEvaluated" | "scopes_evaluated" => Ok(GeneratedField::ScopesEvaluated),
                            "modelVersion" | "model_version" => Ok(GeneratedField::ModelVersion),
                            "validatorAddress" | "validator_address" => Ok(GeneratedField::ValidatorAddress),
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
            type Value = VerificationHistoryEntry;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.VerificationHistoryEntry")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<VerificationHistoryEntry, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut entry_id__ = None;
                let mut timestamp__ = None;
                let mut block_height__ = None;
                let mut previous_score__ = None;
                let mut new_score__ = None;
                let mut previous_status__ = None;
                let mut new_status__ = None;
                let mut scopes_evaluated__ = None;
                let mut model_version__ = None;
                let mut validator_address__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::EntryId => {
                            if entry_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("entryId"));
                            }
                            entry_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Timestamp => {
                            if timestamp__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestamp"));
                            }
                            timestamp__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousScore => {
                            if previous_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousScore"));
                            }
                            previous_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NewScore => {
                            if new_score__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newScore"));
                            }
                            new_score__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PreviousStatus => {
                            if previous_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("previousStatus"));
                            }
                            previous_status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::NewStatus => {
                            if new_status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("newStatus"));
                            }
                            new_status__ = Some(map_.next_value::<AccountStatus>()? as i32);
                        }
                        GeneratedField::ScopesEvaluated => {
                            if scopes_evaluated__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopesEvaluated"));
                            }
                            scopes_evaluated__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ModelVersion => {
                            if model_version__.is_some() {
                                return Err(serde::de::Error::duplicate_field("modelVersion"));
                            }
                            model_version__ = Some(map_.next_value()?);
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
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(VerificationHistoryEntry {
                    entry_id: entry_id__.unwrap_or_default(),
                    timestamp: timestamp__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    previous_score: previous_score__.unwrap_or_default(),
                    new_score: new_score__.unwrap_or_default(),
                    previous_status: previous_status__.unwrap_or_default(),
                    new_status: new_status__.unwrap_or_default(),
                    scopes_evaluated: scopes_evaluated__.unwrap_or_default(),
                    model_version: model_version__.unwrap_or_default(),
                    validator_address: validator_address__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.VerificationHistoryEntry", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for VerificationStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unknown => "VERIFICATION_STATUS_UNKNOWN",
            Self::Pending => "VERIFICATION_STATUS_PENDING",
            Self::InProgress => "VERIFICATION_STATUS_IN_PROGRESS",
            Self::Verified => "VERIFICATION_STATUS_VERIFIED",
            Self::Rejected => "VERIFICATION_STATUS_REJECTED",
            Self::Expired => "VERIFICATION_STATUS_EXPIRED",
            Self::NeedsAdditionalFactor => "VERIFICATION_STATUS_NEEDS_ADDITIONAL_FACTOR",
            Self::AdditionalFactorPending => "VERIFICATION_STATUS_ADDITIONAL_FACTOR_PENDING",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for VerificationStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "VERIFICATION_STATUS_UNKNOWN",
            "VERIFICATION_STATUS_PENDING",
            "VERIFICATION_STATUS_IN_PROGRESS",
            "VERIFICATION_STATUS_VERIFIED",
            "VERIFICATION_STATUS_REJECTED",
            "VERIFICATION_STATUS_EXPIRED",
            "VERIFICATION_STATUS_NEEDS_ADDITIONAL_FACTOR",
            "VERIFICATION_STATUS_ADDITIONAL_FACTOR_PENDING",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = VerificationStatus;

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
                    "VERIFICATION_STATUS_UNKNOWN" => Ok(VerificationStatus::Unknown),
                    "VERIFICATION_STATUS_PENDING" => Ok(VerificationStatus::Pending),
                    "VERIFICATION_STATUS_IN_PROGRESS" => Ok(VerificationStatus::InProgress),
                    "VERIFICATION_STATUS_VERIFIED" => Ok(VerificationStatus::Verified),
                    "VERIFICATION_STATUS_REJECTED" => Ok(VerificationStatus::Rejected),
                    "VERIFICATION_STATUS_EXPIRED" => Ok(VerificationStatus::Expired),
                    "VERIFICATION_STATUS_NEEDS_ADDITIONAL_FACTOR" => Ok(VerificationStatus::NeedsAdditionalFactor),
                    "VERIFICATION_STATUS_ADDITIONAL_FACTOR_PENDING" => Ok(VerificationStatus::AdditionalFactorPending),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for WalletScopeInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.scope_id.is_empty() {
            len += 1;
        }
        if self.scope_type != 0 {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if self.added_at != 0 {
            len += 1;
        }
        if self.verified_at != 0 {
            len += 1;
        }
        if self.expires_at != 0 {
            len += 1;
        }
        if self.consent_granted {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.veid.v1.WalletScopeInfo", len)?;
        if !self.scope_id.is_empty() {
            struct_ser.serialize_field("scopeId", &self.scope_id)?;
        }
        if self.scope_type != 0 {
            let v = ScopeType::try_from(self.scope_type)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.scope_type)))?;
            struct_ser.serialize_field("scopeType", &v)?;
        }
        if self.status != 0 {
            let v = VerificationStatus::try_from(self.status)
                .map_err(|_| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if self.added_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("addedAt", ToString::to_string(&self.added_at).as_str())?;
        }
        if self.verified_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("verifiedAt", ToString::to_string(&self.verified_at).as_str())?;
        }
        if self.expires_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("expiresAt", ToString::to_string(&self.expires_at).as_str())?;
        }
        if self.consent_granted {
            struct_ser.serialize_field("consentGranted", &self.consent_granted)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for WalletScopeInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "scope_id",
            "scopeId",
            "scope_type",
            "scopeType",
            "status",
            "added_at",
            "addedAt",
            "verified_at",
            "verifiedAt",
            "expires_at",
            "expiresAt",
            "consent_granted",
            "consentGranted",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ScopeId,
            ScopeType,
            Status,
            AddedAt,
            VerifiedAt,
            ExpiresAt,
            ConsentGranted,
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
                            "scopeId" | "scope_id" => Ok(GeneratedField::ScopeId),
                            "scopeType" | "scope_type" => Ok(GeneratedField::ScopeType),
                            "status" => Ok(GeneratedField::Status),
                            "addedAt" | "added_at" => Ok(GeneratedField::AddedAt),
                            "verifiedAt" | "verified_at" => Ok(GeneratedField::VerifiedAt),
                            "expiresAt" | "expires_at" => Ok(GeneratedField::ExpiresAt),
                            "consentGranted" | "consent_granted" => Ok(GeneratedField::ConsentGranted),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = WalletScopeInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.veid.v1.WalletScopeInfo")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<WalletScopeInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut scope_id__ = None;
                let mut scope_type__ = None;
                let mut status__ = None;
                let mut added_at__ = None;
                let mut verified_at__ = None;
                let mut expires_at__ = None;
                let mut consent_granted__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ScopeId => {
                            if scope_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeId"));
                            }
                            scope_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ScopeType => {
                            if scope_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("scopeType"));
                            }
                            scope_type__ = Some(map_.next_value::<ScopeType>()? as i32);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map_.next_value::<VerificationStatus>()? as i32);
                        }
                        GeneratedField::AddedAt => {
                            if added_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addedAt"));
                            }
                            added_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::VerifiedAt => {
                            if verified_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("verifiedAt"));
                            }
                            verified_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExpiresAt => {
                            if expires_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expiresAt"));
                            }
                            expires_at__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ConsentGranted => {
                            if consent_granted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("consentGranted"));
                            }
                            consent_granted__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(WalletScopeInfo {
                    scope_id: scope_id__.unwrap_or_default(),
                    scope_type: scope_type__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                    added_at: added_at__.unwrap_or_default(),
                    verified_at: verified_at__.unwrap_or_default(),
                    expires_at: expires_at__.unwrap_or_default(),
                    consent_granted: consent_granted__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.veid.v1.WalletScopeInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for WalletStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Unspecified => "WALLET_STATUS_UNSPECIFIED",
            Self::Active => "WALLET_STATUS_ACTIVE",
            Self::Suspended => "WALLET_STATUS_SUSPENDED",
            Self::Revoked => "WALLET_STATUS_REVOKED",
            Self::Expired => "WALLET_STATUS_EXPIRED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for WalletStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "WALLET_STATUS_UNSPECIFIED",
            "WALLET_STATUS_ACTIVE",
            "WALLET_STATUS_SUSPENDED",
            "WALLET_STATUS_REVOKED",
            "WALLET_STATUS_EXPIRED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = WalletStatus;

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
                    "WALLET_STATUS_UNSPECIFIED" => Ok(WalletStatus::Unspecified),
                    "WALLET_STATUS_ACTIVE" => Ok(WalletStatus::Active),
                    "WALLET_STATUS_SUSPENDED" => Ok(WalletStatus::Suspended),
                    "WALLET_STATUS_REVOKED" => Ok(WalletStatus::Revoked),
                    "WALLET_STATUS_EXPIRED" => Ok(WalletStatus::Expired),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}

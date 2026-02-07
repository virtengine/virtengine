// @generated
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
        if !self.reviews.is_empty() {
            len += 1;
        }
        if self.review_sequence != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.GenesisState", len)?;
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if !self.reviews.is_empty() {
            struct_ser.serialize_field("reviews", &self.reviews)?;
        }
        if self.review_sequence != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reviewSequence", ToString::to_string(&self.review_sequence).as_str())?;
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
            "reviews",
            "review_sequence",
            "reviewSequence",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Params,
            Reviews,
            ReviewSequence,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

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
                            "reviews" => Ok(GeneratedField::Reviews),
                            "reviewSequence" | "review_sequence" => Ok(GeneratedField::ReviewSequence),
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
                formatter.write_str("struct virtengine.review.v1.GenesisState")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<GenesisState, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut params__ = None;
                let mut reviews__ = None;
                let mut review_sequence__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Reviews => {
                            if reviews__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviews"));
                            }
                            reviews__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReviewSequence => {
                            if review_sequence__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewSequence"));
                            }
                            review_sequence__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(GenesisState {
                    params: params__,
                    reviews: reviews__.unwrap_or_default(),
                    review_sequence: review_sequence__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.GenesisState", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeleteReview {
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
        if !self.review_id.is_empty() {
            len += 1;
        }
        if !self.reason.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.MsgDeleteReview", len)?;
        if !self.authority.is_empty() {
            struct_ser.serialize_field("authority", &self.authority)?;
        }
        if !self.review_id.is_empty() {
            struct_ser.serialize_field("reviewId", &self.review_id)?;
        }
        if !self.reason.is_empty() {
            struct_ser.serialize_field("reason", &self.reason)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeleteReview {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "authority",
            "review_id",
            "reviewId",
            "reason",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Authority,
            ReviewId,
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
                            "reviewId" | "review_id" => Ok(GeneratedField::ReviewId),
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
            type Value = MsgDeleteReview;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.review.v1.MsgDeleteReview")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeleteReview, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut authority__ = None;
                let mut review_id__ = None;
                let mut reason__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Authority => {
                            if authority__.is_some() {
                                return Err(serde::de::Error::duplicate_field("authority"));
                            }
                            authority__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReviewId => {
                            if review_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewId"));
                            }
                            review_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reason => {
                            if reason__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reason"));
                            }
                            reason__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgDeleteReview {
                    authority: authority__.unwrap_or_default(),
                    review_id: review_id__.unwrap_or_default(),
                    reason: reason__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.MsgDeleteReview", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgDeleteReviewResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let len = 0;
        let struct_ser = serializer.serialize_struct("virtengine.review.v1.MsgDeleteReviewResponse", len)?;
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgDeleteReviewResponse {
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
            type Value = MsgDeleteReviewResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.review.v1.MsgDeleteReviewResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgDeleteReviewResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                while map_.next_key::<GeneratedField>()?.is_some() {
                    let _ = map_.next_value::<serde::de::IgnoredAny>()?;
                }
                Ok(MsgDeleteReviewResponse {
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.MsgDeleteReviewResponse", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitReview {
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
        if !self.subject_address.is_empty() {
            len += 1;
        }
        if !self.subject_type.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if self.rating != 0 {
            len += 1;
        }
        if !self.comment.is_empty() {
            len += 1;
        }
        if !self.lease_id.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.MsgSubmitReview", len)?;
        if !self.reviewer.is_empty() {
            struct_ser.serialize_field("reviewer", &self.reviewer)?;
        }
        if !self.subject_address.is_empty() {
            struct_ser.serialize_field("subjectAddress", &self.subject_address)?;
        }
        if !self.subject_type.is_empty() {
            struct_ser.serialize_field("subjectType", &self.subject_type)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if self.rating != 0 {
            struct_ser.serialize_field("rating", &self.rating)?;
        }
        if !self.comment.is_empty() {
            struct_ser.serialize_field("comment", &self.comment)?;
        }
        if !self.lease_id.is_empty() {
            struct_ser.serialize_field("leaseId", &self.lease_id)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitReview {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "reviewer",
            "subject_address",
            "subjectAddress",
            "subject_type",
            "subjectType",
            "order_id",
            "orderId",
            "rating",
            "comment",
            "lease_id",
            "leaseId",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Reviewer,
            SubjectAddress,
            SubjectType,
            OrderId,
            Rating,
            Comment,
            LeaseId,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

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
                            "subjectAddress" | "subject_address" => Ok(GeneratedField::SubjectAddress),
                            "subjectType" | "subject_type" => Ok(GeneratedField::SubjectType),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "rating" => Ok(GeneratedField::Rating),
                            "comment" => Ok(GeneratedField::Comment),
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = MsgSubmitReview;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.review.v1.MsgSubmitReview")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitReview, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut reviewer__ = None;
                let mut subject_address__ = None;
                let mut subject_type__ = None;
                let mut order_id__ = None;
                let mut rating__ = None;
                let mut comment__ = None;
                let mut lease_id__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Reviewer => {
                            if reviewer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewer"));
                            }
                            reviewer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubjectAddress => {
                            if subject_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("subjectAddress"));
                            }
                            subject_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubjectType => {
                            if subject_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("subjectType"));
                            }
                            subject_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Rating => {
                            if rating__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rating"));
                            }
                            rating__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Comment => {
                            if comment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("comment"));
                            }
                            comment__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(MsgSubmitReview {
                    reviewer: reviewer__.unwrap_or_default(),
                    subject_address: subject_address__.unwrap_or_default(),
                    subject_type: subject_type__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    rating: rating__.unwrap_or_default(),
                    comment: comment__.unwrap_or_default(),
                    lease_id: lease_id__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.MsgSubmitReview", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for MsgSubmitReviewResponse {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.review_id.is_empty() {
            len += 1;
        }
        if self.submitted_at != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.MsgSubmitReviewResponse", len)?;
        if !self.review_id.is_empty() {
            struct_ser.serialize_field("reviewId", &self.review_id)?;
        }
        if self.submitted_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("submittedAt", ToString::to_string(&self.submitted_at).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for MsgSubmitReviewResponse {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "review_id",
            "reviewId",
            "submitted_at",
            "submittedAt",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReviewId,
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
                            "reviewId" | "review_id" => Ok(GeneratedField::ReviewId),
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
            type Value = MsgSubmitReviewResponse;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.review.v1.MsgSubmitReviewResponse")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<MsgSubmitReviewResponse, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut review_id__ = None;
                let mut submitted_at__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReviewId => {
                            if review_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewId"));
                            }
                            review_id__ = Some(map_.next_value()?);
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
                Ok(MsgSubmitReviewResponse {
                    review_id: review_id__.unwrap_or_default(),
                    submitted_at: submitted_at__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.MsgSubmitReviewResponse", FIELDS, GeneratedVisitor)
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
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.MsgUpdateParams", len)?;
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
                formatter.write_str("struct virtengine.review.v1.MsgUpdateParams")
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
        deserializer.deserialize_struct("virtengine.review.v1.MsgUpdateParams", FIELDS, GeneratedVisitor)
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
        let struct_ser = serializer.serialize_struct("virtengine.review.v1.MsgUpdateParamsResponse", len)?;
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
                formatter.write_str("struct virtengine.review.v1.MsgUpdateParamsResponse")
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
        deserializer.deserialize_struct("virtengine.review.v1.MsgUpdateParamsResponse", FIELDS, GeneratedVisitor)
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
        if self.min_review_interval != 0 {
            len += 1;
        }
        if self.max_comment_length != 0 {
            len += 1;
        }
        if self.require_completed_order {
            len += 1;
        }
        if self.review_window != 0 {
            len += 1;
        }
        if self.min_rating != 0 {
            len += 1;
        }
        if self.max_rating != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.Params", len)?;
        if self.min_review_interval != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("minReviewInterval", ToString::to_string(&self.min_review_interval).as_str())?;
        }
        if self.max_comment_length != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("maxCommentLength", ToString::to_string(&self.max_comment_length).as_str())?;
        }
        if self.require_completed_order {
            struct_ser.serialize_field("requireCompletedOrder", &self.require_completed_order)?;
        }
        if self.review_window != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("reviewWindow", ToString::to_string(&self.review_window).as_str())?;
        }
        if self.min_rating != 0 {
            struct_ser.serialize_field("minRating", &self.min_rating)?;
        }
        if self.max_rating != 0 {
            struct_ser.serialize_field("maxRating", &self.max_rating)?;
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
            "min_review_interval",
            "minReviewInterval",
            "max_comment_length",
            "maxCommentLength",
            "require_completed_order",
            "requireCompletedOrder",
            "review_window",
            "reviewWindow",
            "min_rating",
            "minRating",
            "max_rating",
            "maxRating",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MinReviewInterval,
            MaxCommentLength,
            RequireCompletedOrder,
            ReviewWindow,
            MinRating,
            MaxRating,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "minReviewInterval" | "min_review_interval" => Ok(GeneratedField::MinReviewInterval),
                            "maxCommentLength" | "max_comment_length" => Ok(GeneratedField::MaxCommentLength),
                            "requireCompletedOrder" | "require_completed_order" => Ok(GeneratedField::RequireCompletedOrder),
                            "reviewWindow" | "review_window" => Ok(GeneratedField::ReviewWindow),
                            "minRating" | "min_rating" => Ok(GeneratedField::MinRating),
                            "maxRating" | "max_rating" => Ok(GeneratedField::MaxRating),
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
                formatter.write_str("struct virtengine.review.v1.Params")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Params, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut min_review_interval__ = None;
                let mut max_comment_length__ = None;
                let mut require_completed_order__ = None;
                let mut review_window__ = None;
                let mut min_rating__ = None;
                let mut max_rating__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MinReviewInterval => {
                            if min_review_interval__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minReviewInterval"));
                            }
                            min_review_interval__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxCommentLength => {
                            if max_comment_length__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxCommentLength"));
                            }
                            max_comment_length__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::RequireCompletedOrder => {
                            if require_completed_order__.is_some() {
                                return Err(serde::de::Error::duplicate_field("requireCompletedOrder"));
                            }
                            require_completed_order__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReviewWindow => {
                            if review_window__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewWindow"));
                            }
                            review_window__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MinRating => {
                            if min_rating__.is_some() {
                                return Err(serde::de::Error::duplicate_field("minRating"));
                            }
                            min_rating__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MaxRating => {
                            if max_rating__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxRating"));
                            }
                            max_rating__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Params {
                    min_review_interval: min_review_interval__.unwrap_or_default(),
                    max_comment_length: max_comment_length__.unwrap_or_default(),
                    require_completed_order: require_completed_order__.unwrap_or_default(),
                    review_window: review_window__.unwrap_or_default(),
                    min_rating: min_rating__.unwrap_or_default(),
                    max_rating: max_rating__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.Params", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Review {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.review_id.is_empty() {
            len += 1;
        }
        if !self.reviewer.is_empty() {
            len += 1;
        }
        if !self.subject_address.is_empty() {
            len += 1;
        }
        if !self.subject_type.is_empty() {
            len += 1;
        }
        if !self.order_id.is_empty() {
            len += 1;
        }
        if !self.lease_id.is_empty() {
            len += 1;
        }
        if self.rating != 0 {
            len += 1;
        }
        if !self.comment.is_empty() {
            len += 1;
        }
        if self.submitted_at != 0 {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.review.v1.Review", len)?;
        if !self.review_id.is_empty() {
            struct_ser.serialize_field("reviewId", &self.review_id)?;
        }
        if !self.reviewer.is_empty() {
            struct_ser.serialize_field("reviewer", &self.reviewer)?;
        }
        if !self.subject_address.is_empty() {
            struct_ser.serialize_field("subjectAddress", &self.subject_address)?;
        }
        if !self.subject_type.is_empty() {
            struct_ser.serialize_field("subjectType", &self.subject_type)?;
        }
        if !self.order_id.is_empty() {
            struct_ser.serialize_field("orderId", &self.order_id)?;
        }
        if !self.lease_id.is_empty() {
            struct_ser.serialize_field("leaseId", &self.lease_id)?;
        }
        if self.rating != 0 {
            struct_ser.serialize_field("rating", &self.rating)?;
        }
        if !self.comment.is_empty() {
            struct_ser.serialize_field("comment", &self.comment)?;
        }
        if self.submitted_at != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("submittedAt", ToString::to_string(&self.submitted_at).as_str())?;
        }
        if self.block_height != 0 {
            #[allow(clippy::needless_borrow)]
            #[allow(clippy::needless_borrows_for_generic_args)]
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Review {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "review_id",
            "reviewId",
            "reviewer",
            "subject_address",
            "subjectAddress",
            "subject_type",
            "subjectType",
            "order_id",
            "orderId",
            "lease_id",
            "leaseId",
            "rating",
            "comment",
            "submitted_at",
            "submittedAt",
            "block_height",
            "blockHeight",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ReviewId,
            Reviewer,
            SubjectAddress,
            SubjectType,
            OrderId,
            LeaseId,
            Rating,
            Comment,
            SubmittedAt,
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
                            "reviewId" | "review_id" => Ok(GeneratedField::ReviewId),
                            "reviewer" => Ok(GeneratedField::Reviewer),
                            "subjectAddress" | "subject_address" => Ok(GeneratedField::SubjectAddress),
                            "subjectType" | "subject_type" => Ok(GeneratedField::SubjectType),
                            "orderId" | "order_id" => Ok(GeneratedField::OrderId),
                            "leaseId" | "lease_id" => Ok(GeneratedField::LeaseId),
                            "rating" => Ok(GeneratedField::Rating),
                            "comment" => Ok(GeneratedField::Comment),
                            "submittedAt" | "submitted_at" => Ok(GeneratedField::SubmittedAt),
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
            type Value = Review;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.review.v1.Review")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Review, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut review_id__ = None;
                let mut reviewer__ = None;
                let mut subject_address__ = None;
                let mut subject_type__ = None;
                let mut order_id__ = None;
                let mut lease_id__ = None;
                let mut rating__ = None;
                let mut comment__ = None;
                let mut submitted_at__ = None;
                let mut block_height__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::ReviewId => {
                            if review_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewId"));
                            }
                            review_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Reviewer => {
                            if reviewer__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reviewer"));
                            }
                            reviewer__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubjectAddress => {
                            if subject_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("subjectAddress"));
                            }
                            subject_address__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubjectType => {
                            if subject_type__.is_some() {
                                return Err(serde::de::Error::duplicate_field("subjectType"));
                            }
                            subject_type__ = Some(map_.next_value()?);
                        }
                        GeneratedField::OrderId => {
                            if order_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("orderId"));
                            }
                            order_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::LeaseId => {
                            if lease_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("leaseId"));
                            }
                            lease_id__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Rating => {
                            if rating__.is_some() {
                                return Err(serde::de::Error::duplicate_field("rating"));
                            }
                            rating__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Comment => {
                            if comment__.is_some() {
                                return Err(serde::de::Error::duplicate_field("comment"));
                            }
                            comment__ = Some(map_.next_value()?);
                        }
                        GeneratedField::SubmittedAt => {
                            if submitted_at__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submittedAt"));
                            }
                            submitted_at__ = 
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
                Ok(Review {
                    review_id: review_id__.unwrap_or_default(),
                    reviewer: reviewer__.unwrap_or_default(),
                    subject_address: subject_address__.unwrap_or_default(),
                    subject_type: subject_type__.unwrap_or_default(),
                    order_id: order_id__.unwrap_or_default(),
                    lease_id: lease_id__.unwrap_or_default(),
                    rating: rating__.unwrap_or_default(),
                    comment: comment__.unwrap_or_default(),
                    submitted_at: submitted_at__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.review.v1.Review", FIELDS, GeneratedVisitor)
    }
}

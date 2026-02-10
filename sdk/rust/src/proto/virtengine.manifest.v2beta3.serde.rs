// @generated
impl serde::Serialize for Group {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.services.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.Group", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.services.is_empty() {
            struct_ser.serialize_field("services", &self.services)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Group {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "services",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Services,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "name" => Ok(GeneratedField::Name),
                            "services" => Ok(GeneratedField::Services),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Group;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.Group")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Group, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut services__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Services => {
                            if services__.is_some() {
                                return Err(serde::de::Error::duplicate_field("services"));
                            }
                            services__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(Group {
                    name: name__.unwrap_or_default(),
                    services: services__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.Group", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ImageCredentials {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.host.is_empty() {
            len += 1;
        }
        if !self.email.is_empty() {
            len += 1;
        }
        if !self.username.is_empty() {
            len += 1;
        }
        if !self.password.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.ImageCredentials", len)?;
        if !self.host.is_empty() {
            struct_ser.serialize_field("host", &self.host)?;
        }
        if !self.email.is_empty() {
            struct_ser.serialize_field("email", &self.email)?;
        }
        if !self.username.is_empty() {
            struct_ser.serialize_field("username", &self.username)?;
        }
        if !self.password.is_empty() {
            struct_ser.serialize_field("password", &self.password)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ImageCredentials {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "host",
            "email",
            "username",
            "password",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Host,
            Email,
            Username,
            Password,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "host" => Ok(GeneratedField::Host),
                            "email" => Ok(GeneratedField::Email),
                            "username" => Ok(GeneratedField::Username),
                            "password" => Ok(GeneratedField::Password),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ImageCredentials;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.ImageCredentials")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ImageCredentials, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut host__ = None;
                let mut email__ = None;
                let mut username__ = None;
                let mut password__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Host => {
                            if host__.is_some() {
                                return Err(serde::de::Error::duplicate_field("host"));
                            }
                            host__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Email => {
                            if email__.is_some() {
                                return Err(serde::de::Error::duplicate_field("email"));
                            }
                            email__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Username => {
                            if username__.is_some() {
                                return Err(serde::de::Error::duplicate_field("username"));
                            }
                            username__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Password => {
                            if password__.is_some() {
                                return Err(serde::de::Error::duplicate_field("password"));
                            }
                            password__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ImageCredentials {
                    host: host__.unwrap_or_default(),
                    email: email__.unwrap_or_default(),
                    username: username__.unwrap_or_default(),
                    password: password__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.ImageCredentials", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Service {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.image.is_empty() {
            len += 1;
        }
        if !self.command.is_empty() {
            len += 1;
        }
        if !self.args.is_empty() {
            len += 1;
        }
        if !self.env.is_empty() {
            len += 1;
        }
        if self.resources.is_some() {
            len += 1;
        }
        if self.count != 0 {
            len += 1;
        }
        if !self.expose.is_empty() {
            len += 1;
        }
        if self.params.is_some() {
            len += 1;
        }
        if self.credentials.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.Service", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.image.is_empty() {
            struct_ser.serialize_field("image", &self.image)?;
        }
        if !self.command.is_empty() {
            struct_ser.serialize_field("command", &self.command)?;
        }
        if !self.args.is_empty() {
            struct_ser.serialize_field("args", &self.args)?;
        }
        if !self.env.is_empty() {
            struct_ser.serialize_field("env", &self.env)?;
        }
        if let Some(v) = self.resources.as_ref() {
            struct_ser.serialize_field("resources", v)?;
        }
        if self.count != 0 {
            struct_ser.serialize_field("count", &self.count)?;
        }
        if !self.expose.is_empty() {
            struct_ser.serialize_field("expose", &self.expose)?;
        }
        if let Some(v) = self.params.as_ref() {
            struct_ser.serialize_field("params", v)?;
        }
        if let Some(v) = self.credentials.as_ref() {
            struct_ser.serialize_field("credentials", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Service {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "image",
            "command",
            "args",
            "env",
            "resources",
            "count",
            "expose",
            "params",
            "credentials",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Image,
            Command,
            Args,
            Env,
            Resources,
            Count,
            Expose,
            Params,
            Credentials,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "name" => Ok(GeneratedField::Name),
                            "image" => Ok(GeneratedField::Image),
                            "command" => Ok(GeneratedField::Command),
                            "args" => Ok(GeneratedField::Args),
                            "env" => Ok(GeneratedField::Env),
                            "resources" => Ok(GeneratedField::Resources),
                            "count" => Ok(GeneratedField::Count),
                            "expose" => Ok(GeneratedField::Expose),
                            "params" => Ok(GeneratedField::Params),
                            "credentials" => Ok(GeneratedField::Credentials),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Service;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.Service")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<Service, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut image__ = None;
                let mut command__ = None;
                let mut args__ = None;
                let mut env__ = None;
                let mut resources__ = None;
                let mut count__ = None;
                let mut expose__ = None;
                let mut params__ = None;
                let mut credentials__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Image => {
                            if image__.is_some() {
                                return Err(serde::de::Error::duplicate_field("image"));
                            }
                            image__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Command => {
                            if command__.is_some() {
                                return Err(serde::de::Error::duplicate_field("command"));
                            }
                            command__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Args => {
                            if args__.is_some() {
                                return Err(serde::de::Error::duplicate_field("args"));
                            }
                            args__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Env => {
                            if env__.is_some() {
                                return Err(serde::de::Error::duplicate_field("env"));
                            }
                            env__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Resources => {
                            if resources__.is_some() {
                                return Err(serde::de::Error::duplicate_field("resources"));
                            }
                            resources__ = map_.next_value()?;
                        }
                        GeneratedField::Count => {
                            if count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("count"));
                            }
                            count__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Expose => {
                            if expose__.is_some() {
                                return Err(serde::de::Error::duplicate_field("expose"));
                            }
                            expose__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Params => {
                            if params__.is_some() {
                                return Err(serde::de::Error::duplicate_field("params"));
                            }
                            params__ = map_.next_value()?;
                        }
                        GeneratedField::Credentials => {
                            if credentials__.is_some() {
                                return Err(serde::de::Error::duplicate_field("credentials"));
                            }
                            credentials__ = map_.next_value()?;
                        }
                    }
                }
                Ok(Service {
                    name: name__.unwrap_or_default(),
                    image: image__.unwrap_or_default(),
                    command: command__.unwrap_or_default(),
                    args: args__.unwrap_or_default(),
                    env: env__.unwrap_or_default(),
                    resources: resources__,
                    count: count__.unwrap_or_default(),
                    expose: expose__.unwrap_or_default(),
                    params: params__,
                    credentials: credentials__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.Service", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceExpose {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.port != 0 {
            len += 1;
        }
        if self.external_port != 0 {
            len += 1;
        }
        if !self.proto.is_empty() {
            len += 1;
        }
        if !self.service.is_empty() {
            len += 1;
        }
        if self.global {
            len += 1;
        }
        if !self.hosts.is_empty() {
            len += 1;
        }
        if self.http_options.is_some() {
            len += 1;
        }
        if !self.ip.is_empty() {
            len += 1;
        }
        if self.endpoint_sequence_number != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.ServiceExpose", len)?;
        if self.port != 0 {
            struct_ser.serialize_field("port", &self.port)?;
        }
        if self.external_port != 0 {
            struct_ser.serialize_field("externalPort", &self.external_port)?;
        }
        if !self.proto.is_empty() {
            struct_ser.serialize_field("proto", &self.proto)?;
        }
        if !self.service.is_empty() {
            struct_ser.serialize_field("service", &self.service)?;
        }
        if self.global {
            struct_ser.serialize_field("global", &self.global)?;
        }
        if !self.hosts.is_empty() {
            struct_ser.serialize_field("hosts", &self.hosts)?;
        }
        if let Some(v) = self.http_options.as_ref() {
            struct_ser.serialize_field("httpOptions", v)?;
        }
        if !self.ip.is_empty() {
            struct_ser.serialize_field("ip", &self.ip)?;
        }
        if self.endpoint_sequence_number != 0 {
            struct_ser.serialize_field("endpointSequenceNumber", &self.endpoint_sequence_number)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceExpose {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "port",
            "external_port",
            "externalPort",
            "proto",
            "service",
            "global",
            "hosts",
            "http_options",
            "httpOptions",
            "ip",
            "endpoint_sequence_number",
            "endpointSequenceNumber",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Port,
            ExternalPort,
            Proto,
            Service,
            Global,
            Hosts,
            HttpOptions,
            Ip,
            EndpointSequenceNumber,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "port" => Ok(GeneratedField::Port),
                            "externalPort" | "external_port" => Ok(GeneratedField::ExternalPort),
                            "proto" => Ok(GeneratedField::Proto),
                            "service" => Ok(GeneratedField::Service),
                            "global" => Ok(GeneratedField::Global),
                            "hosts" => Ok(GeneratedField::Hosts),
                            "httpOptions" | "http_options" => Ok(GeneratedField::HttpOptions),
                            "ip" => Ok(GeneratedField::Ip),
                            "endpointSequenceNumber" | "endpoint_sequence_number" => Ok(GeneratedField::EndpointSequenceNumber),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceExpose;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.ServiceExpose")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceExpose, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut port__ = None;
                let mut external_port__ = None;
                let mut proto__ = None;
                let mut service__ = None;
                let mut global__ = None;
                let mut hosts__ = None;
                let mut http_options__ = None;
                let mut ip__ = None;
                let mut endpoint_sequence_number__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Port => {
                            if port__.is_some() {
                                return Err(serde::de::Error::duplicate_field("port"));
                            }
                            port__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ExternalPort => {
                            if external_port__.is_some() {
                                return Err(serde::de::Error::duplicate_field("externalPort"));
                            }
                            external_port__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Proto => {
                            if proto__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proto"));
                            }
                            proto__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Service => {
                            if service__.is_some() {
                                return Err(serde::de::Error::duplicate_field("service"));
                            }
                            service__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Global => {
                            if global__.is_some() {
                                return Err(serde::de::Error::duplicate_field("global"));
                            }
                            global__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Hosts => {
                            if hosts__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hosts"));
                            }
                            hosts__ = Some(map_.next_value()?);
                        }
                        GeneratedField::HttpOptions => {
                            if http_options__.is_some() {
                                return Err(serde::de::Error::duplicate_field("httpOptions"));
                            }
                            http_options__ = map_.next_value()?;
                        }
                        GeneratedField::Ip => {
                            if ip__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ip"));
                            }
                            ip__ = Some(map_.next_value()?);
                        }
                        GeneratedField::EndpointSequenceNumber => {
                            if endpoint_sequence_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("endpointSequenceNumber"));
                            }
                            endpoint_sequence_number__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ServiceExpose {
                    port: port__.unwrap_or_default(),
                    external_port: external_port__.unwrap_or_default(),
                    proto: proto__.unwrap_or_default(),
                    service: service__.unwrap_or_default(),
                    global: global__.unwrap_or_default(),
                    hosts: hosts__.unwrap_or_default(),
                    http_options: http_options__,
                    ip: ip__.unwrap_or_default(),
                    endpoint_sequence_number: endpoint_sequence_number__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.ServiceExpose", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceExposeHttpOptions {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.max_body_size != 0 {
            len += 1;
        }
        if self.read_timeout != 0 {
            len += 1;
        }
        if self.send_timeout != 0 {
            len += 1;
        }
        if self.next_tries != 0 {
            len += 1;
        }
        if self.next_timeout != 0 {
            len += 1;
        }
        if !self.next_cases.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.ServiceExposeHTTPOptions", len)?;
        if self.max_body_size != 0 {
            struct_ser.serialize_field("maxBodySize", &self.max_body_size)?;
        }
        if self.read_timeout != 0 {
            struct_ser.serialize_field("readTimeout", &self.read_timeout)?;
        }
        if self.send_timeout != 0 {
            struct_ser.serialize_field("sendTimeout", &self.send_timeout)?;
        }
        if self.next_tries != 0 {
            struct_ser.serialize_field("nextTries", &self.next_tries)?;
        }
        if self.next_timeout != 0 {
            struct_ser.serialize_field("nextTimeout", &self.next_timeout)?;
        }
        if !self.next_cases.is_empty() {
            struct_ser.serialize_field("nextCases", &self.next_cases)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceExposeHttpOptions {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "max_body_size",
            "maxBodySize",
            "read_timeout",
            "readTimeout",
            "send_timeout",
            "sendTimeout",
            "next_tries",
            "nextTries",
            "next_timeout",
            "nextTimeout",
            "next_cases",
            "nextCases",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            MaxBodySize,
            ReadTimeout,
            SendTimeout,
            NextTries,
            NextTimeout,
            NextCases,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "maxBodySize" | "max_body_size" => Ok(GeneratedField::MaxBodySize),
                            "readTimeout" | "read_timeout" => Ok(GeneratedField::ReadTimeout),
                            "sendTimeout" | "send_timeout" => Ok(GeneratedField::SendTimeout),
                            "nextTries" | "next_tries" => Ok(GeneratedField::NextTries),
                            "nextTimeout" | "next_timeout" => Ok(GeneratedField::NextTimeout),
                            "nextCases" | "next_cases" => Ok(GeneratedField::NextCases),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceExposeHttpOptions;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.ServiceExposeHTTPOptions")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceExposeHttpOptions, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut max_body_size__ = None;
                let mut read_timeout__ = None;
                let mut send_timeout__ = None;
                let mut next_tries__ = None;
                let mut next_timeout__ = None;
                let mut next_cases__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::MaxBodySize => {
                            if max_body_size__.is_some() {
                                return Err(serde::de::Error::duplicate_field("maxBodySize"));
                            }
                            max_body_size__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ReadTimeout => {
                            if read_timeout__.is_some() {
                                return Err(serde::de::Error::duplicate_field("readTimeout"));
                            }
                            read_timeout__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SendTimeout => {
                            if send_timeout__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sendTimeout"));
                            }
                            send_timeout__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NextTries => {
                            if next_tries__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nextTries"));
                            }
                            next_tries__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NextTimeout => {
                            if next_timeout__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nextTimeout"));
                            }
                            next_timeout__ = 
                                Some(map_.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::NextCases => {
                            if next_cases__.is_some() {
                                return Err(serde::de::Error::duplicate_field("nextCases"));
                            }
                            next_cases__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(ServiceExposeHttpOptions {
                    max_body_size: max_body_size__.unwrap_or_default(),
                    read_timeout: read_timeout__.unwrap_or_default(),
                    send_timeout: send_timeout__.unwrap_or_default(),
                    next_tries: next_tries__.unwrap_or_default(),
                    next_timeout: next_timeout__.unwrap_or_default(),
                    next_cases: next_cases__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.ServiceExposeHTTPOptions", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ServiceParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.storage.is_empty() {
            len += 1;
        }
        if self.credentials.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.ServiceParams", len)?;
        if !self.storage.is_empty() {
            struct_ser.serialize_field("storage", &self.storage)?;
        }
        if let Some(v) = self.credentials.as_ref() {
            struct_ser.serialize_field("credentials", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ServiceParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "storage",
            "credentials",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Storage,
            Credentials,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "storage" => Ok(GeneratedField::Storage),
                            "credentials" => Ok(GeneratedField::Credentials),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ServiceParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.ServiceParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<ServiceParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut storage__ = None;
                let mut credentials__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Storage => {
                            if storage__.is_some() {
                                return Err(serde::de::Error::duplicate_field("storage"));
                            }
                            storage__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Credentials => {
                            if credentials__.is_some() {
                                return Err(serde::de::Error::duplicate_field("credentials"));
                            }
                            credentials__ = map_.next_value()?;
                        }
                    }
                }
                Ok(ServiceParams {
                    storage: storage__.unwrap_or_default(),
                    credentials: credentials__,
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.ServiceParams", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for StorageParams {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.name.is_empty() {
            len += 1;
        }
        if !self.mount.is_empty() {
            len += 1;
        }
        if self.read_only {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("virtengine.manifest.v2beta3.StorageParams", len)?;
        if !self.name.is_empty() {
            struct_ser.serialize_field("name", &self.name)?;
        }
        if !self.mount.is_empty() {
            struct_ser.serialize_field("mount", &self.mount)?;
        }
        if self.read_only {
            struct_ser.serialize_field("readOnly", &self.read_only)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for StorageParams {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "name",
            "mount",
            "read_only",
            "readOnly",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Name,
            Mount,
            ReadOnly,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "name" => Ok(GeneratedField::Name),
                            "mount" => Ok(GeneratedField::Mount),
                            "readOnly" | "read_only" => Ok(GeneratedField::ReadOnly),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = StorageParams;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct virtengine.manifest.v2beta3.StorageParams")
            }

            fn visit_map<V>(self, mut map_: V) -> std::result::Result<StorageParams, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut name__ = None;
                let mut mount__ = None;
                let mut read_only__ = None;
                while let Some(k) = map_.next_key()? {
                    match k {
                        GeneratedField::Name => {
                            if name__.is_some() {
                                return Err(serde::de::Error::duplicate_field("name"));
                            }
                            name__ = Some(map_.next_value()?);
                        }
                        GeneratedField::Mount => {
                            if mount__.is_some() {
                                return Err(serde::de::Error::duplicate_field("mount"));
                            }
                            mount__ = Some(map_.next_value()?);
                        }
                        GeneratedField::ReadOnly => {
                            if read_only__.is_some() {
                                return Err(serde::de::Error::duplicate_field("readOnly"));
                            }
                            read_only__ = Some(map_.next_value()?);
                        }
                    }
                }
                Ok(StorageParams {
                    name: name__.unwrap_or_default(),
                    mount: mount__.unwrap_or_default(),
                    read_only: read_only__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("virtengine.manifest.v2beta3.StorageParams", FIELDS, GeneratedVisitor)
    }
}

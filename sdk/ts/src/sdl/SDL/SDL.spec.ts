import fs from "node:fs";
import path from "node:path";

import { faker } from "@faker-js/faker";
import { beforeEach, describe, expect, it } from "@jest/globals";
import { createGroups, createManifest, createSdlJson, createSdlYml } from "@test/helpers/sdl";

import { VE_DENOM, SANDBOX_ID, USDC_IBC_DENOMS } from "../../network/config.ts";
import type { v2ServiceImageCredentials } from "../types.ts";
import { SDL } from "./SDL.ts";
import { SdlValidationError } from "./SdlValidationError.ts";

describe("SDL", () => {
  describe("profiles placement pricing denomination", () => {
    it.each([VE_DENOM, USDC_IBC_DENOMS[SANDBOX_ID]])("should resolve a group with a valid \"%s\" denomination", (denom) => {
      const sdl = SDL.fromString(
        createSdlYml({
          "profiles.placement.dcloud.pricing.web.denom": { $set: denom },
        }),
        "beta3",
        "sandbox",
      );

      expect(sdl.manifest()).toMatchObject(createManifest());
      expect(sdl.groups()).toMatchObject(
        createGroups({
          "0.resources.0.price.denom": { $set: denom },
        }),
      );
    });

    it("should throw an error when denomination is invalid", () => {
      const denom = "usdt";
      const yml = createSdlYml({
        "profiles.placement.dcloud.pricing.web.denom": { $set: denom },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrow(
        new SdlValidationError("Invalid format: \"denom\" at \"/profiles/placement/dcloud/pricing/web\" does not match pattern \"^(uve|ibc/.*)$\""),
      );

      const anotherYaml = createSdlYml({
        "profiles.placement.dcloud.pricing.web.denom": { $set: "ibc/1234567890" },
      });

      expect(() => SDL.fromString(anotherYaml, "beta3", "sandbox")).toThrow(
        new SdlValidationError(`Invalid denom: "ibc/1234567890" at path "/profiles/placement/dcloud/pricing/web/denom". Only "uve" and "ibc/12C6A0C374171B595A0A9E18B83FA09D295FB1F2D8C6DAA3AC28683471752D84" are supported.`),
      );
    });
  });

  describe("endpoints", () => {
    it("should resolve with valid endpoints", () => {
      const endpointName = faker.lorem.word({ length: { min: 3, max: 10 } });
      const endpoint = {
        [endpointName]: {
          kind: "ip",
        },
      };
      const yml = createSdlYml({
        endpoints: { $set: endpoint },
        "services.web.expose.0.to.0.ip": { $set: endpointName },
      });
      const sdl = SDL.fromString(yml, "beta3", "sandbox");

      expect(sdl.manifest()).toMatchObject(
        createManifest({
          "0.services.0.resources.endpoints.1": {
            $set: {
              kind: 2,
              sequence_number: 1,
            },
          },
          "0.services.0.expose": {
            $set: [
              {
                ip: endpointName,
                endpointSequenceNumber: 1,
              },
            ],
          },
        }),
      );
      expect(sdl.groups()).toMatchObject(
        createGroups({
          "0.resources.0.resource.endpoints": {
            $push: [
              {
                kind: 2,
                sequence_number: 1,
              },
            ],
          },
        }),
      );
    });

    it("should throw provided an invalid endpoint name", () => {
      const endpointName = faker.number.int().toString();
      const endpoint = {
        [endpointName]: {
          kind: "ip",
        },
      };
      const yml = createSdlYml({
        endpoints: { $set: endpoint },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError(
        `Field "${endpointName}" at "/endpoints" doesn't satisfy any of the allowed patterns: ^[a-z]+[-_0-9a-z]+$.`,
      ));
    });

    it("should throw provided no endpoint kind", () => {
      const endpointName = faker.lorem.word({ length: { min: 3, max: 10 } });
      const endpoint = {
        [endpointName]: {},
      };
      const yml = createSdlYml({
        endpoints: { $set: endpoint },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError(`Missing required field: "kind" at "/endpoints/${endpointName}".`));
    });

    it("should throw provided invalid endpoint kind", () => {
      const endpointName = faker.lorem.word({ length: { min: 3, max: 10 } });
      const endpointKind = faker.lorem.word();
      const endpoint = {
        [endpointName]: {
          kind: endpointKind,
        },
      };
      const yml = createSdlYml({
        endpoints: { $set: endpoint },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError(`"kind" at "/endpoints/${endpointName}" should be one of: ip.`),
      );
    });

    it("should throw when endpoint is unused", () => {
      const endpointName = faker.lorem.word({ length: { min: 3, max: 10 } });
      const endpoint = {
        [endpointName]: {
          kind: "ip",
        },
      };
      const yml = createSdlYml({
        endpoints: { $set: endpoint },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError(`Endpoint "${endpointName}" declared but never used.`));
    });
  });

  describe("service image credentials", () => {
    it("should resolve a service with valid credentials", () => {
      const credentials = {
        host: faker.internet.url(),
        username: faker.internet.username(),
        password: faker.internet.password(),
      };
      const yml = createSdlYml({
        "services.web.credentials": { $set: credentials },
      });
      const sdl = SDL.fromString(yml, "beta3", "sandbox");

      expect(sdl.manifest()).toMatchObject(
        createManifest({
          "0.services.0.credentials": {
            $set: {
              ...credentials,
              email: "",
            },
          },
        }),
      );
      expect(sdl.groups()).toMatchObject(createGroups());
    });

    it("should resolve a service without credentials", () => {
      const sdl = SDL.fromString(createSdlYml({}), "beta3", "sandbox");

      expect(sdl.manifest()).toMatchObject(
        createManifest({
          "0.services.0.credentials": {
            $set: null,
          },
        }),
      );
      expect(sdl.groups()).toMatchObject(createGroups());
    });

    describe("invalid credentials errors", () => {
      const fields: (keyof v2ServiceImageCredentials)[] = ["host", "username", "password"];
      let credentials: v2ServiceImageCredentials;

      beforeEach(() => {
        credentials = {
          host: faker.internet.url(),
          username: faker.internet.username(),
          password: faker.internet.password(),
        };
      });

      it.each(fields)("should throw an error when credentials are missing \"%s\"", (field) => {
        const { [field]: _, ...credentialsWithoutField } = credentials;
        const yml = createSdlYml({
          "services.web.credentials": { $set: credentialsWithoutField },
        });

        expect(() => {
          SDL.fromString(yml, "beta3", "sandbox");
        }).toThrowError(new SdlValidationError(`Missing required field: "${field}" at "/services/web/credentials".`));
      });

      it.each(fields)("should throw an error when credentials \"%s\" is empty", (field) => {
        credentials[field] = "";
        const yml = createSdlYml({
          "services.web.credentials": { $set: credentials },
        });

        expect(() => {
          SDL.fromString(yml, "beta3", "sandbox");
        }).toThrowError(new RegExp(`"${field}" at "/services/web/credentials" must be at least \\d+ characters long`));
      });
    });
  });

  describe("deployment validation", () => {
    it("should throw an error when a service is not defined in deployment", () => {
      const yml = createSdlYml({
        deployment: { $unset: ["web"] },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError("Service \"web\" is not defined at \"/deployment\" section."));
    });

    it("should throw an error when deployment is not defined in profile placement", () => {
      const yml = createSdlYml({
        "profiles.placement": { $unset: ["dcloud"] },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("The placement \"dcloud\" is not defined in the \"placement\" section."),
      );
    });

    it("should throw an error when service compute requirements are not defined", () => {
      const yml = createSdlYml({
        "profiles.compute": { $unset: ["web"] },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("The compute requirements for the \"web\" profile are not defined in the \"compute\" section."),
      );
    });
  });

  describe("storage validation", () => {
    it("should throw an error when a service references a non-existing volume", () => {
      const yml = createSdlYml({
        "services.web.params": { $set: { storage: { data: { mount: "/mnt/data", readOnly: false } } } },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("Service \"web\" references non-existing compute volume \"data\"."),
      );
    });

    it("should throw an error when a service volume mount is a non-absolute path", () => {
      const yml = createSdlYml({
        "services.web.params": { $set: { storage: { data: { mount: "./mnt/data", readOnly: false } } } },
        "profiles.compute.web.resources.storage": { $set: { name: "data", size: "1Gi" } },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError(`Invalid format: "mount" at "/services/web/params/storage/data" does not match pattern "^/"`),
      );
    });

    it("should throw an error when multiple ephemeral storages are provided", () => {
      const yml = createSdlYml({
        "services.web.params": {
          $set: {
            storage: {
              data: {},
              db: {},
            },
          },
        },
        "profiles.compute.web.resources.storage": {
          $set: [
            { name: "data", size: "1Gi" },
            { name: "db", size: "1Gi" },
          ],
        },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError("Multiple root ephemeral storages are not allowed."));
    });

    it("should throw an error when mount is used by multiple volumes", () => {
      const yml = createSdlYml({
        "services.web.params": { $set: { storage: { data: { mount: "/", readOnly: false }, db: { mount: "/", readOnly: false } } } },
        "profiles.compute.web.resources.storage": {
          $set: [
            { name: "data", size: "1Gi" },
            { name: "db", size: "1Gi" },
          ],
        },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError("Mount \"/\" already in use by volume \"data\"."));
    });

    it("should require a service storage mount if volume is persistent", () => {
      const yml = createSdlYml({
        "services.web.params": {
          $set: { storage: { data: { readOnly: false } } },
        },
        "profiles.compute.web.resources.storage": { $set: { name: "data", size: "1Gi", attributes: { persistent: true } } },
      });

      expect(() => SDL.fromString(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("Persistent storage \"data\" requires a mount path in /services/web/params/storage/data/mount."),
      );
    });

    it("should throw an error when ram storage is defined as persistent", () => {
      const yml = createSdlJson({
        "profiles.compute.web.resources.storage": { $set: { name: "data", size: "1Gi", attributes: { class: "ram", persistent: true } } },
      });
      expect(() => new SDL(yml, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("\"ram\" storage at \"/profiles/compute/web/resources/storage\" cannot be persistent"),
      );
    });

    it("should throw an error if storage size in not defined", () => {
      const yml = createSdlJson({
        "profiles.compute.web.resources.storage": { $set: { name: "data" } },
      });
      expect(() => new SDL(yml, "beta3", "sandbox")).toThrowError(new SdlValidationError("Missing required field: \"size\" at \"/profiles/compute/web/resources/storage\"."));
    });
  });

  describe("gpu validation", () => {
    it("should throw an error when gpu units is not defined", () => {
      const sdlJson = createSdlJson({
        "profiles.compute.web.resources.gpu": { $set: {} },
      });

      expect(() => new SDL(sdlJson, "beta3", "sandbox")).toThrowError(new SdlValidationError("Missing required field: \"units\" at \"/profiles/compute/web/resources/gpu\"."));
    });

    it("should throw an error when gpu units > 0 and attributes is not defined", () => {
      const sdlJson = createSdlJson({
        "profiles.compute.web.resources.gpu": { $set: { units: 1 } },
      });

      expect(() => new SDL(sdlJson, "beta3", "sandbox")).toThrowError(new SdlValidationError("GPU must have attributes if units is not 0."));
    });

    it("should throw an error when gpu units is 0 and attributes is defined", () => {
      const sdlJson = createSdlJson({
        "profiles.compute.web.resources.gpu": { $set: { units: 0, attributes: {} } },
      });

      expect(() => new SDL(sdlJson, "beta3", "sandbox")).toThrowError(new SdlValidationError("GPU must not have attributes if units is 0."));
    });

    it("should throw an error when gpu units > 0 and attributes vendor is not supported", () => {
      const sdlJson = createSdlJson({
        "profiles.compute.web.resources.gpu": {
          $set: {
            units: 1,
            attributes: {
              vendor: {
                nvidia: { model: "rtxa6000" },
              },
            },
          },
        },
      });

      expect(() => new SDL(sdlJson, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("\"nvidia\" at \"/profiles/compute/web/resources/gpu/attributes/vendor\" should be array."),
      );
    });

    it("should throw an error when gpu units > 0 and attributes vendor is not supported", () => {
      const sdlJson = createSdlJson({
        "profiles.compute.web.resources.gpu": {
          $set: {
            units: 1,
            attributes: {
              vendor: {
                nvidia: [{ model: "rtxa6000", interface: "foo" }],
              },
            },
          },
        },
      });

      expect(() => new SDL(sdlJson, "beta3", "sandbox")).toThrowError(
        new SdlValidationError("\"interface\" at \"/profiles/compute/web/resources/gpu/attributes/vendor/nvidia/0\" should be one of: pcie, sxm."),
      );
    });
  });

  describe("test sdl persistent storage", () => {
    it("SDL: Persistent Storage Manifest", () => {
      const validSDL = readFileSync("./fixtures/persistent_storage_valid.sdl.yaml");

      const sdl = SDL.fromString(validSDL, "beta2");
      const result = sdl.manifest();

      expect(result).toMatchSnapshot("SDL: Persistent Storage Manifest");
    });
  });

  describe("test GPU with interface", () => {
    const testSDL = readFileSync("./fixtures/gpu_basic_ram_interface.sdl.yaml");

    const expectedManifest = JSON.parse(readFileSync("./fixtures/gpu_basic_ram_interface.manifest.json"));

    it("SDL: GPU Manifest with ram & interface", () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = sdl.manifest();

      expect(result).toStrictEqual(expectedManifest);
    });

    it("SDL: GPU Version with ram & interface (snapshot)", async () => {
      const sdl = SDL.fromString(testSDL, "beta3");

      await expect(sdl.manifestVersion()).resolves.toMatchSnapshot("Manifest version matches expected result");
    });
  });

  describe("SDL GPU Invalid Vendor", () => {
    it("SDL: GPU must throw if the vendor is invalid", () => {
      const invalidSDL = readFileSync("./fixtures/gpu_invalid_vendor.sdl.yaml");

      const t = () => {
        SDL.fromString(invalidSDL, "beta3");
      };

      expect(t).toThrow(`Additional property "invalidvendor" is not allowed at "/profiles/compute/web/resources/gpu/attributes/vendor".`);
    });

    it("SDL: GPU without vendor name should throw", () => {
      const invalidSDL = readFileSync("./fixtures/gpu_invalid_no_vendor_name.sdl.yaml");

      const t = () => {
        SDL.fromString(invalidSDL, "beta3");
      };

      expect(t).toThrow("\"vendor\" at \"/profiles/compute/web/resources/gpu/attributes\" must have at least 1 property.");
    });
  });

  describe("SDL WordPress", () => {
    const testSDL = readFileSync("./fixtures/wordpress.sdl.yaml");
    const expectedManifest = JSON.parse(readFileSync("./fixtures/wordpress.manifest.json"));

    describe("Manifest", () => {
      it("should generate the correct manifest", () => {
        const sdl = SDL.fromString(testSDL, "beta3");
        const result = sdl.manifestSorted();

        expect(result).toEqual(expectedManifest);
      });
    });

    describe("Deployment Groups", () => {
      it("should generate the correct deployment groups", () => {
        const sdl = SDL.fromString(testSDL, "beta3");
        const result = sdl.groups();

        expect(result).toMatchSnapshot();
      });
    });

    describe("Version", () => {
      it("should return the correct manifest version", async () => {
        const sdl = SDL.fromString(testSDL, "beta3");
        const result = await sdl.manifestVersion();

        expect(result).toMatchSnapshot();
      });
    });
  });

  describe("SDL v3 Resource Groups", () => {
    it("should create v3 resource groups", async () => {
      const testSDL = readFileSync("./fixtures/gpu_basic.sdl.yaml");
      const sdl = SDL.fromString(testSDL, "beta3");

      expect(sdl.groups()).toMatchSnapshot("Groups matches expected result");
    });
  });

  describe("SDL Minesweeper", () => {
    const testSDL = `
    version: '2.0'
    services:
      minesweeper:
        image: creepto/minesweeper
        expose:
          - port: 3000
            as: 80
            to:
              - global: true
    profiles:
      compute:
        minesweeper:
          resources:
            cpu:
              units: 0.1
            memory:
              size: 512Mi
            storage:
              - size: 512Mi
      placement:
        virtengine:
          attributes:
            organization: virtengine.com
          signedBy:
            anyOf:
              - virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63
              - virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4
          pricing:
            minesweeper:
              denom: uve
              amount: 10000
    deployment:
      minesweeper:
        virtengine:
          profile: minesweeper
          count: 1
    `.replace(/^ {4}/gm, "");

    const expectedManifest = [
      {
        Name: "virtengine",
        Services: [
          {
            Name: "minesweeper",
            Image: "creepto/minesweeper",
            Command: null,
            Args: null,
            Env: null,
            Resources: {
              cpu: {
                units: {
                  val: "100",
                },
              },
              memory: {
                quantity: {
                  val: "536870912",
                },
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: "536870912",
                  },
                },
              ],
              endpoints: null,
            },
            Count: 1,
            Expose: [
              {
                Port: 3000,
                ExternalPort: 80,
                Proto: "TCP",
                Service: "",
                Global: true,
                Hosts: null,
                HTTPOptions: {
                  MaxBodySize: 1048576,
                  ReadTimeout: 60000,
                  SendTimeout: 60000,
                  NextTries: 3,
                  NextTimeout: 0,
                  NextCases: ["error", "timeout"],
                },
                IP: "",
                EndpointSequenceNumber: 0,
              },
            ],
          },
        ],
      },
    ];

    const expectedGroups = [
      {
        name: "virtengine",
        requirements: {
          attributes: [
            {
              key: "organization",
              value: "virtengine.com",
            },
          ],
          signedBy: {
            allOf: [],
            anyOf: ["virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63", "virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4"],
          },
        },
        resources: [
          {
            resources: {
              cpu: {
                units: {
                  val: {
                    0: 49,
                    1: 48,
                    2: 48,
                  },
                },
              },
              memory: {
                quantity: {
                  val: {
                    0: 53,
                    1: 51,
                    2: 54,
                    3: 56,
                    4: 55,
                    5: 48,
                    6: 57,
                    7: 49,
                    8: 50,
                  },
                },
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: {
                      0: 53,
                      1: 51,
                      2: 54,
                      3: 56,
                      4: 55,
                      5: 48,
                      6: 57,
                      7: 49,
                      8: 50,
                    },
                  },
                },
              ],
              endpoints: [
                {
                  kind: 0,
                  sequence_number: 0,
                },
              ],
            },
            price: {
              denom: "uve",
              amount: "10000",
            },
            count: 1,
          },
        ],
      },
    ];

    const expectedVersion = new Uint8Array([
      117, 11, 114, 73, 243, 231, 14, 234, 211, 32, 100, 158, 202, 240, 89, 251, 6, 222, 2, 248, 30, 169, 146, 97, 176, 103, 44, 196, 64, 220, 97, 102,
    ]);

    describe("fromString", () => {
      it("should create an SDL instance with data", () => {
        const sdl = SDL.fromString(testSDL, "beta2");

        expect(sdl).toBeInstanceOf(SDL);
        expect(sdl.data).not.toBeNull();
      });
    });

    describe("Manifest", () => {
      it("should generate the correct manifest", () => {
        const sdl = SDL.fromString(testSDL, "beta2");
        const result = sdl.manifest(true);

        expect(result).toEqual(expectedManifest);
      });
    });

    describe("DeploymentGroups", () => {
      it("should generate the correct deployment groups", () => {
        const sdl = SDL.fromString(testSDL, "beta2");
        const result = JSON.parse(JSON.stringify(sdl.groups()));

        expect(result).toEqual(expectedGroups);
      });
    });

    describe("Version", () => {
      it("should return the correct manifest version", async () => {
        const sdl = SDL.fromString(testSDL, "beta2");
        const result = await sdl.manifestVersion();

        expect(result).toEqual(expectedVersion);
      });
    });
  });

  describe("SDL: IP Lease Manifest", () => {
    it("should generate the correct manifest", async () => {
      const validSDL = readFileSync("./fixtures/ip_lease_valid.sdl.yaml");

      const sdl = SDL.fromString(validSDL, "beta2");
      const result = sdl.manifest();

      expect(result).toMatchSnapshot("Manifest matches expected result");
    });
  });

  describe("SDL: Manifest w/ HTTP options", () => {
    const testSDL = `
    version: '2.0'
    services:
      tetris:
        image: bsord/tetris
        expose:
          - port: 80
            http_options:
              max_body_size: 104857600
              read_timeout: 50
              send_timeout: 100
              next_tries: 24
              next_timeout: 48
              next_cases:
                - "500"
            as: 80
            to:
              - global: true
    profiles:
      compute:
        tetris:
          resources:
            cpu:
              units: 1
            memory:
              size: 512Mi
            storage:
              - size: 512Mi
      placement:
        virtengine:
          attributes:
            host: virtengine
          signedBy:
            anyOf:
              - virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63
              - virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4
          pricing:
            tetris:
              denom: uve
              amount: 10000
    deployment:
      tetris:
        virtengine:
          profile: tetris
          count: 1
    `.replace(/^ {4}/gm, "");

    const testManifest = [
      {
        Name: "virtengine",
        Services: [
          {
            Name: "tetris",
            Image: "bsord/tetris",
            Command: null,
            Args: null,
            Env: null,
            Resources: {
              cpu: {
                units: {
                  val: "1000",
                },
              },
              memory: {
                quantity: {
                  val: "536870912",
                },
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: "536870912",
                  },
                },
              ],
              endpoints: null,
            },
            Count: 1,
            Expose: [
              {
                Port: 80,
                ExternalPort: 80,
                Proto: "TCP",
                Service: "",
                Global: true,
                Hosts: null,
                HTTPOptions: {
                  MaxBodySize: 104857600,
                  ReadTimeout: 50,
                  SendTimeout: 100,
                  NextTries: 24,
                  NextTimeout: 48,
                  NextCases: ["500"],
                },
                IP: "",
                EndpointSequenceNumber: 0,
              },
            ],
          },
        ],
      },
    ];

    it("should generate the correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const result = sdl.manifest(true);

      expect(result).toEqual(testManifest);
    });
  });

  describe("SDL GPU", () => {
    const testSDL = readFileSync("./fixtures/gpu_basic.sdl.yaml");
    const expectedManifest = JSON.parse(readFileSync("./fixtures/gpu_basic.manifest.json"));

    it("should generate the correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = sdl.manifest(true);

      expect(result).toEqual(expectedManifest);
    });

    it("should return the correct manifest version", async () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = await sdl.manifestVersion();

      expect(result).toMatchSnapshot("Manifest version matches expected result");
    });
  });

  describe("SDL GPU Tests", () => {
    const testSDL = `
    version: "2.0"
    services:
      web:
        image: nginx
        expose:
          - port: 80
            accept:
              - ahostname.com
            to:
              - global: true
          - port: 12345
            to:
              - global: true
            proto: udp
    profiles:
      compute:
        web:
          resources:
            cpu:
              units: "100m"
            gpu:
              units: 1
              attributes:
                vendor:
                  nvidia:
            memory:
              size: "128Mi"
            storage:
            - size: "1Gi"
      placement:
        westcoast:
          attributes:
            region: us-west
            blalbla: foo
          signedBy:
            anyOf:
              - virt1123123
              - virt1124123123
            allOf:
              - virt1123123
              - virt1124123123
          pricing:
            web:
              denom: uve
              amount: 50
    deployment:
      web:
        westcoast:
          profile: web
          count: 2
    `.replace(/^ {4}/gm, "");

    it("should generate correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = sdl.manifest();

      expect(result).toMatchSnapshot();
    });

    it("should return correct version", async () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = await sdl.manifestVersion();

      expect(result).toMatchSnapshot();
    });
  });

  describe("SDL GPU RAM", () => {
    const testSDL = readFileSync("./fixtures/gpu_basic_ram.sdl.yaml");
    const expectedManifest = JSON.parse(readFileSync("./fixtures/gpu_basic_ram.manifest.json"));

    it("should generate correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = sdl.manifest(true);

      expect(result).toEqual(expectedManifest);
    });

    it("should return correct version", async () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = await sdl.manifestVersion();

      expect(result).toMatchSnapshot();
    });
  });

  describe("SDL: fromString", () => {
    const validSDL = readFileSync("./fixtures/gpu_no_gpu_valid.sdl.yaml");
    const hasAttrSDL = readFileSync("./fixtures/gpu_no_gpu_invalid_has_attributes.sdl.yaml");
    const noVendorSdl = readFileSync("./fixtures/gpu_invalid_no_vendor.sdl.yaml");
    const invalidIntefaceSdl = readFileSync("./fixtures/gpu_invalid_interface.sdl.yaml");

    it("should accept if GPU units is 0, and no attributes are present", () => {
      expect(() => SDL.fromString(validSDL, "beta3")).not.toThrow();
    });

    it("should throw an error if GPU units is not 0, and there are no attributes present", () => {
      expect(() => SDL.fromString(hasAttrSDL, "beta3")).toThrow(
        /GPU must not have attributes if units is 0/,
      );
    });

    it("should throw an error if GPU units is not 0, and the vendor is not present", () => {
      expect(() => SDL.fromString(noVendorSdl, "beta3")).toThrow(
        /GPU must specify a vendor if units is not 0/,
      );
    });

    it("should throw an error if GPU interface is not supported", () => {
      expect(() => SDL.fromString(invalidIntefaceSdl, "beta3")).toThrow(
        /"interface" at "\/profiles\/compute\/web\/resources\/gpu\/attributes\/vendor\/nvidia\/0" should be one of: pcie, sxm\./,
      );
    });
  });

  describe("SDL: Manifest w/ env", () => {
    const testSDL = `
    version: '2.0'
    services:
      tetris:
        image: bsord/tetris
        env:
          - EMPTY=
          - ENV1=test1
          - ENV2=test2
        expose:
          - port: 80
            as: 80
            to:
              - global: true
    profiles:
      compute:
        tetris:
          resources:
            cpu:
              units: 1
            memory:
              size: 512Mi
            storage:
              - size: 512Mi
      placement:
        virtengine:
          attributes:
            host: virtengine
          signedBy:
            anyOf:
              - virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63
              - virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4
          pricing:
            tetris:
              denom: uve
              amount: 10000
    deployment:
      tetris:
        virtengine:
          profile: tetris
          count: 1
    `.replace(/^ {4}/gm, "");

    const testManifest = [
      {
        Name: "virtengine",
        Services: [
          {
            Name: "tetris",
            Image: "bsord/tetris",
            Command: null,
            Args: null,
            Env: ["EMPTY=", "ENV1=test1", "ENV2=test2"],
            Resources: {
              cpu: {
                units: {
                  val: "1000",
                },
              },
              memory: {
                quantity: {
                  val: "536870912",
                },
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: "536870912",
                  },
                },
              ],
              endpoints: null,
            },
            Count: 1,
            Expose: [
              {
                Port: 80,
                ExternalPort: 80,
                Proto: "TCP",
                Service: "",
                Global: true,
                Hosts: null,
                HTTPOptions: {
                  MaxBodySize: 1048576,
                  ReadTimeout: 60000,
                  SendTimeout: 60000,
                  NextTries: 3,
                  NextTimeout: 0,
                  NextCases: ["error", "timeout"],
                },
                IP: "",
                EndpointSequenceNumber: 0,
              },
            ],
          },
        ],
      },
    ];

    it("should generate correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const result = sdl.manifest(true);

      expect(result).toEqual(testManifest);
    });
  });

  describe("SDL: Empty Profile", () => {
    const testSDL = `
    version: '2.0'
    services:
      tetris-main:
        image: bsord/tetris
        expose:
          - port: 80
            as: 80
            to:
              - global: true
    profiles:
      compute:
        tetris:
          resources:
            cpu:
              units: 1
            memory:
              size: 512Mi
            storage:
              - size: 512Mi
      placement:
        virtengine:
          attributes:
            host: virtengine
          signedBy:
            anyOf:
              - virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63
              - virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4
          pricing:
            tetris:
              denom: uve
              amount: 10000
    deployment:
      tetris-main:
        virtengine:
          profile: tetris
          count: 1
    `.replace(/^ {4}/gm, "");

    const expectedManifest = [
      {
        Name: "virtengine",
        Services: [
          {
            Name: "tetris-main",
            Image: "bsord/tetris",
            Command: null,
            Args: null,
            Env: null,
            Resources: {
              cpu: {
                units: {
                  val: "1000",
                },
              },
              memory: {
                quantity: {
                  val: "536870912",
                },
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: "536870912",
                  },
                },
              ],
              endpoints: null,
            },
            Count: 1,
            Expose: [
              {
                Port: 80,
                ExternalPort: 80,
                Proto: "TCP",
                Service: "",
                Global: true,
                Hosts: null,
                HTTPOptions: {
                  MaxBodySize: 1048576,
                  ReadTimeout: 60000,
                  SendTimeout: 60000,
                  NextTries: 3,
                  NextTimeout: 0,
                  NextCases: ["error", "timeout"],
                },
                IP: "",
                EndpointSequenceNumber: 0,
              },
            ],
          },
        ],
      },
    ];

    const expectedPreVersionJson
      = "[{\"Name\":\"virtengine\",\"Services\":[{\"Args\":null,\"Command\":null,\"Count\":1,\"Env\":null,\"Expose\":[{\"EndpointSequenceNumber\":0,\"ExternalPort\":80,\"Global\":true,\"HTTPOptions\":{\"MaxBodySize\":1048576,\"NextCases\":[\"error\",\"timeout\"],\"NextTimeout\":0,\"NextTries\":3,\"ReadTimeout\":60000,\"SendTimeout\":60000},\"Hosts\":null,\"IP\":\"\",\"Port\":80,\"Proto\":\"TCP\",\"Service\":\"\"}],\"Image\":\"bsord/tetris\",\"Name\":\"tetris-main\",\"Resources\":{\"cpu\":{\"units\":{\"val\":\"1000\"}},\"endpoints\":null,\"memory\":{\"size\":{\"val\":\"536870912\"}},\"storage\":[{\"name\":\"default\",\"size\":{\"val\":\"536870912\"}}]}}]}]";

    const expectedVersion = new Uint8Array([
      247, 77, 26, 95, 231, 205, 208, 76, 208, 217, 59, 106, 109, 76, 73, 196, 37, 14, 75, 170, 210, 120, 231, 213, 69, 226, 219, 203, 236, 116, 106, 135,
    ]);

    it("should create SDL instance from string", () => {
      const sdl = SDL.fromString(testSDL, "beta2");

      expect(sdl).toBeInstanceOf(SDL);
      expect(sdl.data).not.toBeNull();
    });

    it("should generate correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const result = sdl.manifest(true);

      expect(result).toEqual(expectedManifest);
    });

    it("should generate correct manifest version", async () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const preversionJson = sdl.manifestSortedJSON();
      const result = await sdl.manifestVersion();

      expect(preversionJson).toBe(expectedPreVersionJson);
      expect(result).toEqual(expectedVersion);
    });
  });

  describe("SDL: Basic", () => {
    const testSDL = `
    version: '2.0'
    services:
      tetris:
        image: bsord/tetris
        expose:
          - port: 80
            as: 80
            to:
              - global: true
    profiles:
      compute:
        tetris:
          resources:
            cpu:
              units: 1
            memory:
              size: 512Mi
            storage:
              - size: 512Mi
      placement:
        virtengine:
          attributes:
            host: virtengine
          signedBy:
            anyOf:
              - virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63
              - virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4
          pricing:
            tetris:
              denom: uve
              amount: 10000
    deployment:
      tetris:
        virtengine:
          profile: tetris
          count: 1
    `.replace(/^ {4}/gm, "");

    const expectedManifest = [
      {
        Name: "virtengine",
        Services: [
          {
            Name: "tetris",
            Image: "bsord/tetris",
            Command: null,
            Args: null,
            Env: null,
            Resources: {
              cpu: {
                units: {
                  val: "1000",
                },
              },
              memory: {
                quantity: {
                  val: "536870912",
                },
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: "536870912",
                  },
                },
              ],
              endpoints: null,
            },
            Count: 1,
            Expose: [
              {
                Port: 80,
                ExternalPort: 80,
                Proto: "TCP",
                Service: "",
                Global: true,
                Hosts: null,
                HTTPOptions: {
                  MaxBodySize: 1048576,
                  ReadTimeout: 60000,
                  SendTimeout: 60000,
                  NextTries: 3,
                  NextTimeout: 0,
                  NextCases: ["error", "timeout"],
                },
                IP: "",
                EndpointSequenceNumber: 0,
              },
            ],
          },
        ],
      },
    ];

    const expectedGroups = [
      {
        name: "virtengine",
        requirements: {
          attributes: [
            {
              key: "host",
              value: "virtengine",
            },
          ],
          signedBy: {
            allOf: [],
            anyOf: ["virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63", "virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4"],
          },
        },
        resources: [
          {
            resources: {
              cpu: {
                units: {
                  val: new Uint8Array([49, 48, 48, 48]),
                },
                attributes: undefined,
              },
              memory: {
                quantity: {
                  val: new Uint8Array([53, 51, 54, 56, 55, 48, 57, 49, 50]),
                },
                attributes: undefined,
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: new Uint8Array([53, 51, 54, 56, 55, 48, 57, 49, 50]),
                  },
                  attributes: undefined,
                },
              ],
              endpoints: [
                {
                  kind: 0,
                  sequence_number: 0,
                },
              ],
            },
            price: {
              denom: "uve",
              amount: "10000",
            },
            count: 1,
          },
        ],
      },
    ];

    const expectedPreVersionJson
      = "[{\"Name\":\"virtengine\",\"Services\":[{\"Args\":null,\"Command\":null,\"Count\":1,\"Env\":null,\"Expose\":[{\"EndpointSequenceNumber\":0,\"ExternalPort\":80,\"Global\":true,\"HTTPOptions\":{\"MaxBodySize\":1048576,\"NextCases\":[\"error\",\"timeout\"],\"NextTimeout\":0,\"NextTries\":3,\"ReadTimeout\":60000,\"SendTimeout\":60000},\"Hosts\":null,\"IP\":\"\",\"Port\":80,\"Proto\":\"TCP\",\"Service\":\"\"}],\"Image\":\"bsord/tetris\",\"Name\":\"tetris\",\"Resources\":{\"cpu\":{\"units\":{\"val\":\"1000\"}},\"endpoints\":null,\"memory\":{\"size\":{\"val\":\"536870912\"}},\"storage\":[{\"name\":\"default\",\"size\":{\"val\":\"536870912\"}}]}}]}]";

    const expectedVersion = new Uint8Array([
      119, 216, 99, 142, 214, 87, 92, 43, 168, 216, 126, 234, 231, 181, 75, 175, 102, 90, 218, 232, 182, 53, 158, 91, 6, 246, 3, 15, 63, 120, 212, 176,
    ]);

    it("should create SDL instance from string", () => {
      const sdl = SDL.fromString(testSDL, "beta2");

      expect(sdl).toBeInstanceOf(SDL);
      expect(sdl.data).not.toBeNull();
    });

    it("should generate correct manifest", () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const result = sdl.manifest(true);

      expect(result).toEqual(expectedManifest);
    });

    it("should generate correct deployment groups", () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const result = sdl.groups();

      expect(result).toEqual(expectedGroups);
    });

    it("should generate correct manifest version", async () => {
      const sdl = SDL.fromString(testSDL, "beta2");
      const preversionJson = sdl.manifestSortedJSON();
      const result = await sdl.manifestVersion();

      expect(preversionJson).toBe(expectedPreVersionJson);
      expect(result).toEqual(expectedVersion);
    });
  });

  describe("SDL: Basic Beta3", () => {
    const testSDL = `
    version: '2.0'
    services:
      tetris:
        image: bsord/tetris
        expose:
          - port: 80
            as: 80
            to:
              - global: true
    profiles:
      compute:
        tetris:
          resources:
            cpu:
              units: 1
            gpu:
              units: 0
            memory:
              size: 512Mi
            storage:
              - size: 512Mi
      placement:
        virtengine:
          attributes:
            host: virtengine
          signedBy:
            anyOf:
              - virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63
              - virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4
          pricing:
            tetris:
              denom: uve
              amount: 10000
    deployment:
      tetris:
        virtengine:
          profile: tetris
          count: 1
    `.replace(/^ {4}/gm, "");

    const expectedGroups = [
      {
        name: "virtengine",
        requirements: {
          attributes: [
            {
              key: "host",
              value: "virtengine",
            },
          ],
          signedBy: {
            allOf: [],
            anyOf: ["virt1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63", "virt18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4"],
          },
        },
        resources: [
          {
            resource: {
              cpu: {
                units: {
                  val: new Uint8Array([49, 48, 48, 48]),
                },
                attributes: undefined,
              },
              memory: {
                quantity: {
                  val: new Uint8Array([53, 51, 54, 56, 55, 48, 57, 49, 50]),
                },
                attributes: undefined,
              },
              storage: [
                {
                  name: "default",
                  quantity: {
                    val: new Uint8Array([53, 51, 54, 56, 55, 48, 57, 49, 50]),
                  },
                  attributes: undefined,
                },
              ],
              gpu: {
                units: {
                  val: new Uint8Array([48]),
                },
                attributes: undefined,
              },
              endpoints: [
                {
                  sequence_number: 0,
                },
              ],
              id: 1,
            },
            price: {
              denom: "uve",
              amount: "10000",
            },
            count: 1,
          },
        ],
      },
    ];

    it("should create SDL instance from string", () => {
      const sdl = SDL.fromString(testSDL, "beta3");

      expect(sdl).toBeInstanceOf(SDL);
      expect(sdl.data).not.toBeNull();
    });

    it("should generate correct deployment groups", () => {
      const sdl = SDL.fromString(testSDL, "beta3");
      const result = sdl.groups();

      expect(result).toEqual(expectedGroups);
    });
  });

  /**
   * @param {FilePath} filePath Relative path from current directory
   */
  function readFileSync(filePath: string) {
    const fullPath = path.join(__dirname, filePath);
    return fs.readFileSync(fullPath, "utf8");
  }
});

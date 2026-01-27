import { describe, expect, it } from "@jest/globals";
import { merge } from "lodash";

import type { DeepPartial } from "../../../encoding/typeEncodingHelpers.ts";
import type { NetworkId } from "../../../network/index.ts";
import { AKT_DENOM, USDC_IBC_DENOMS } from "../../../network/index.ts";
import { type SDLInput, validateSDL } from "./validateSDL.ts";

describe(validateSDL.name, () => {
  describe("valid SDL", () => {
    it("returns undefined for a valid SDL", () => {
      const { validate } = setup();
      expect(validate()).toBeUndefined();
    });

    it("returns undefined for valid SDL with USDC denom on sandbox", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: {
                  amount: "1000",
                  denom: USDC_IBC_DENOMS.sandbox,
                },
              },
            },
          },
        },
      }, "sandbox");
      expect(validate()).toBeUndefined();
    });

    it("returns undefined for valid SDL with USDC denom on mainnet", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: {
                  amount: "1000",
                  denom: USDC_IBC_DENOMS.mainnet,
                },
              },
            },
          },
        },
      }, "mainnet");
      expect(validate()).toBeUndefined();
    });
  });

  describe("denom validation", () => {
    it("returns an error for invalid denom", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: {
                  amount: "1000",
                  denom: "ibc/invalid",
                },
              },
            },
          },
        },
      }, "sandbox");

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("Invalid denom: \"ibc/invalid\""),
        instancePath: "/profiles/placement/dcloud/pricing/web/denom",
        schemaPath: "#/definitions/priceCoin/properties/denom",
        keyword: "pattern",
      }));
    });

    it("returns an error when using sandbox USDC on mainnet", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: {
                  amount: "1000",
                  denom: USDC_IBC_DENOMS.sandbox,
                },
              },
            },
          },
        },
      }, "mainnet");

      const errors = validate();
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining(`Invalid denom: "${USDC_IBC_DENOMS.sandbox}"`),
        keyword: "pattern",
      }));
    });
  });

  describe("deployment validation", () => {
    it("returns an error when service is not defined in deployment", () => {
      const sdl: SDLInput = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          // web is missing here
          other: {
            dcloud: { count: 1, profile: "web" },
          },
        },
      };

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Service \"web\" is not defined at \"/deployment\" section.",
        instancePath: "/deployment",
        schemaPath: "#/properties/deployment",
        keyword: "required",
        params: { missingProperty: "web" },
      }));
    });

    it("returns an error when placement is not defined", () => {
      const { validate } = setup({
        deployment: {
          web: {
            "nonexistent-placement": {
              count: 1,
              profile: "web",
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "The placement \"nonexistent-placement\" is not defined in the \"placement\" section.",
        instancePath: "/profiles/placement",
        schemaPath: "#/properties/profiles/properties/placement",
        keyword: "required",
        params: { missingProperty: "nonexistent-placement" },
      }));
    });

    it("returns an error when compute profile is not defined", () => {
      const { validate } = setup({
        deployment: {
          web: {
            dcloud: {
              count: 1,
              profile: "nonexistent-profile",
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "The compute requirements for the \"nonexistent-profile\" profile are not defined in the \"compute\" section.",
        instancePath: "/profiles/compute",
        schemaPath: "#/properties/profiles/properties/compute",
        keyword: "required",
        params: { missingProperty: "nonexistent-profile" },
      }));
    });

    it("returns an error when pricing for profile is not defined", () => {
      const sdl: SDLInput = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                other: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: {
            dcloud: { count: 1, profile: "web" },
          },
        },
      };

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: "The pricing for the \"web\" profile is not defined in the \"dcloud\" placement.",
        instancePath: "/profiles/placement/dcloud/pricing",
        schemaPath: "#/properties/profiles/properties/placement/additionalProperties/properties/pricing",
        keyword: "required",
        params: { missingProperty: "web" },
      }));
    });
  });

  describe("storage validation", () => {
    it("returns an error when service references non-existing storage volume", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: { mount: "/data" },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Service \"web\" references non-existing compute volume \"data\".",
        instancePath: "/profiles/compute/web/resources/storage",
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/storage",
        keyword: "required",
        params: { missingProperty: "data" },
      }));
    });

    it("returns an error for multiple root ephemeral storages", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: {},
                logs: {},
              },
            },
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: [
                  { name: "data", size: "1Gi" },
                  { name: "logs", size: "1Gi" },
                ],
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Multiple root ephemeral storages are not allowed.",
        schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage",
        keyword: "uniqueItems",
      }));
    });

    it("returns an error when mount is used by multiple volumes", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: { mount: "/mnt" },
                logs: { mount: "/mnt" },
              },
            },
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: [
                  { name: "data", size: "1Gi" },
                  { name: "logs", size: "1Gi" },
                ],
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Mount \"/mnt\" already in use by volume \"data\".",
        instancePath: "/services/web/params/storage/logs/mount",
        schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage/additionalProperties/properties/mount",
        keyword: "uniqueItems",
        params: { duplicate: "/mnt" },
      }));
    });

    it("returns an error when persistent storage has no mount", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: { readOnly: false },
              },
            },
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: {
                  name: "data",
                  size: "1Gi",
                  attributes: { persistent: true },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Persistent storage \"data\" requires a mount path in /services/web/params/storage/data/mount.",
        instancePath: "/services/web/params/storage/data",
        schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage/additionalProperties/properties/mount",
        keyword: "required",
        params: { missingProperty: "mount" },
      }));
    });

    it("accepts persistent storage with mount defined", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: { mount: "/data", readOnly: false },
              },
            },
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: {
                  name: "data",
                  size: "1Gi",
                  attributes: { persistent: true },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("GPU validation", () => {
    it("returns an error when GPU units > 0 but no attributes", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: { units: 1 },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "GPU must have attributes if units is not 0.",
        instancePath: "/profiles/compute/web/resources/gpu",
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/gpu",
        keyword: "required",
        params: { missingProperty: "attributes" },
      }));
    });

    it("returns an error when GPU units > 0 but no vendor", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: { units: 1, attributes: {} },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "GPU must specify a vendor if units is not 0.",
        instancePath: "/profiles/compute/web/resources/gpu/attributes",
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/gpu/properties/attributes/properties/vendor",
        keyword: "required",
        params: { missingProperty: "vendor" },
      }));
    });

    it("returns an error when GPU units = 0 but has attributes", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: { units: 0, attributes: { vendor: { nvidia: [] } } },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "GPU must not have attributes if units is 0.",
        instancePath: "/profiles/compute/web/resources/gpu/attributes",
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/gpu/properties/attributes",
        keyword: "additionalProperties",
        params: { additionalProperty: "attributes" },
      }));
    });

    it("accepts GPU with units > 0 and valid vendor", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: { units: 1, attributes: { vendor: { nvidia: [] } } },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("accepts GPU with units = 0 and no attributes", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: { units: 0 },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("IP lease validation", () => {
    it("returns an error when IP is declared but not global", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                to: [{ ip: "myendpoint", global: false }],
              },
            ],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        instancePath: "/services/web/expose/0/to/0/global",
        keyword: "const",
        message: "If an IP is declared, the directive must be declared as global.",
        params: { allowedValue: true },
        schemaPath: "#/definitions/exposeToWithIpEnforcesGlobal/then/properties/global/const",
      }));
    });

    it("returns an error when IP references unknown endpoint", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                to: [{ ip: "unknown-endpoint", global: true }],
              },
            ],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Unknown endpoint \"unknown-endpoint\" for service \"web\". Add it to the \"endpoints\" section.",
        instancePath: "/endpoints/unknown-endpoint",
        schemaPath: "#/properties/endpoints",
        keyword: "required",
        params: { missingProperty: "unknown-endpoint" },
      }));
    });

    it("returns an error when same IP endpoint port is used by multiple services", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
          api: {
            image: "node:latest",
            expose: [
              {
                port: 3000,
                as: 80,
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
            api: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
                api: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: {
            dcloud: { count: 1, profile: "web" },
          },
          api: {
            dcloud: { count: 1, profile: "api" },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("IP endpoint \"myendpoint\" port"),
        keyword: "uniqueItems",
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("already in use by"),
      }));
    });

    it("accepts valid IP lease configuration", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("endpoint validation", () => {
    it("returns an error when endpoint is declared but never used", () => {
      const { validate } = setup({
        endpoints: {
          "unused-endpoint": { kind: "ip" },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Endpoint \"unused-endpoint\" declared but never used.",
        instancePath: "/endpoints/unused-endpoint",
        schemaPath: "#/properties/endpoints",
        keyword: "additionalProperties",
        params: { additionalProperty: "unused-endpoint" },
      }));
    });

    it("does not return an error when all endpoints are used", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("multiple services validation", () => {
    it("validates all services in the SDL", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
          api: {
            image: "node:latest",
            expose: [{ port: 3000, as: 3000, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
            api: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
                api: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: {
            dcloud: { count: 1, profile: "web" },
          },
          api: {
            dcloud: { count: 1, profile: "api" },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("returns errors for multiple invalid services", () => {
      const sdl: SDLInput = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
          api: {
            image: "node:latest",
            expose: [{ port: 3000, as: 3000, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
            api: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
                api: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          // both web and api are missing
          other: {
            dcloud: { count: 1, profile: "web" },
          },
        },
      };

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: "Service \"web\" is not defined at \"/deployment\" section.",
        keyword: "required",
        params: { missingProperty: "web" },
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: "Service \"api\" is not defined at \"/deployment\" section.",
        keyword: "required",
        params: { missingProperty: "api" },
      }));
    });
  });

  describe("protocol handling", () => {
    it("handles uppercase TCP protocol", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                proto: "TCP",
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("handles lowercase tcp protocol", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                proto: "tcp",
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("handles UDP protocol", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 53,
                as: 53,
                proto: "UDP",
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("allows same port with different protocols on same IP", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [
              {
                port: 80,
                as: 80,
                proto: "TCP",
                to: [{ ip: "myendpoint", global: true }],
              },
              {
                port: 80,
                as: 80,
                proto: "UDP",
                to: [{ ip: "myendpoint", global: true }],
              },
            ],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("schema validation: services", () => {
    it("returns an error when service image is empty", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"image\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("at least 1 character"),
      }));
    });

    it("returns an error for invalid port number (0)", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 0 as number, as: 80, to: [{ global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"port\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("at least 1"),
      }));
    });

    it("returns an error for invalid port number (65536)", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 65536 as number, as: 80, to: [{ global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"port\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("at most 65535"),
      }));
    });

    it("returns an error for invalid protocol", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, proto: "HTTP" as "TCP", to: [{ global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"proto\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("TCP, UDP, tcp, udp"),
      }));
    });
  });

  describe("schema validation: credentials", () => {
    it("returns an error when credentials host is missing", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            credentials: {
              username: "user",
              password: "password123",
            } as SDLInput["services"][string]["credentials"],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"host\""),
      }));
    });

    it("returns an error when credentials username is missing", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            credentials: {
              host: "registry.example.com",
              password: "password123",
            } as SDLInput["services"][string]["credentials"],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"username\""),
      }));
    });

    it("returns an error when credentials password is missing", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            credentials: {
              host: "registry.example.com",
              username: "user",
            } as SDLInput["services"][string]["credentials"],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"password\""),
      }));
    });

    it("returns an error when credentials password is too short", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            credentials: {
              host: "registry.example.com",
              username: "user",
              password: "short",
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"password\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("at least 6 characters"),
      }));
    });

    it("accepts valid credentials", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            credentials: {
              host: "registry.example.com",
              username: "user",
              password: "password123",
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("schema validation: http_options", () => {
    it("returns an error when max_body_size exceeds 100MB", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{
              port: 80,
              as: 80,
              to: [{ global: true }],
              http_options: {
                max_body_size: 104857601,
              },
            }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"max_body_size\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("104857600"),
      }));
    });

    it("returns an error when read_timeout exceeds 60000ms", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{
              port: 80,
              as: 80,
              to: [{ global: true }],
              http_options: {
                read_timeout: 60001,
              },
            }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"read_timeout\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("60000"),
      }));
    });

    it("returns an error when send_timeout exceeds 60000ms", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{
              port: 80,
              as: 80,
              to: [{ global: true }],
              http_options: {
                send_timeout: 60001,
              },
            }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"send_timeout\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("60000"),
      }));
    });

    it("accepts valid http_options", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{
              port: 80,
              as: 80,
              to: [{ global: true }],
              http_options: {
                max_body_size: 104857600,
                read_timeout: 60000,
                send_timeout: 60000,
                next_tries: 3,
                next_timeout: 0,
                next_cases: ["error", "timeout"],
              },
            }],
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("schema validation: storage mount path", () => {
    it("returns an error when mount path is not absolute", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: { mount: "relative/path" as `/relative/path` },
              },
            },
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { name: "data", size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"mount\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("pattern \"^/\""),
      }));
    });

    it("accepts absolute mount path", () => {
      const { validate } = setup({
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
            params: {
              storage: {
                data: { mount: "/data" },
              },
            },
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { name: "data", size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("schema validation: endpoints", () => {
    it("returns an error for invalid endpoint name pattern", () => {
      const { validate } = setup({
        endpoints: {
          "123invalid": { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ ip: "123invalid", global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"123invalid\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("^[a-z]+[-_0-9a-z]+$"),
      }));
    });

    it("returns an error when endpoint kind is missing", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: {} as { kind: "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ ip: "myendpoint", global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"kind\""),
      }));
    });

    it("returns an error for invalid endpoint kind", () => {
      const { validate } = setup({
        endpoints: {
          myendpoint: { kind: "dns" as "ip" },
        },
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ ip: "myendpoint", global: true }] }],
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"kind\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("ip"),
      }));
    });
  });

  describe("schema validation: profiles.compute.resources", () => {
    it("returns an error when cpu is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"cpu\""),
      }));
    });

    it("returns an error when memory is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"memory\""),
      }));
    });

    it("returns an error when storage is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"storage\""),
      }));
    });

    it("returns an error when cpu.units is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: {},
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"units\""),
      }));
    });

    it("returns an error when memory.size is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: {},
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"size\""),
      }));
    });

    it("returns an error when storage.size is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: {},
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"size\""),
      }));
    });
  });

  describe("schema validation: GPU attributes", () => {
    it("returns an error for invalid GPU interface", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: {
                  units: 1,
                  attributes: {
                    vendor: {
                      nvidia: [{ interface: "invalid" as "pcie" }],
                    },
                  },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"interface\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("pcie, sxm"),
      }));
    });

    it("accepts valid GPU with pcie interface", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: {
                  units: 1,
                  attributes: {
                    vendor: {
                      nvidia: [{ model: "a100", interface: "pcie" }],
                    },
                  },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("accepts valid GPU with sxm interface", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: {
                  units: 1,
                  attributes: {
                    vendor: {
                      nvidia: [{ model: "a100", interface: "sxm" }],
                    },
                  },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });

    it("returns an error for invalid GPU vendor", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
                gpu: {
                  units: 1,
                  attributes: {
                    vendor: {
                      amd: [],
                    },
                  },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"amd\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("not allowed"),
      }));
    });
  });

  describe("schema validation: storage attributes", () => {
    it("returns an error when RAM storage is persistent", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: {
                  name: "data",
                  size: "1Gi",
                  attributes: { class: "ram", persistent: true },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"ram\" storage"),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("cannot be persistent"),
      }));
    });

    it("returns an error when RAM storage persistent is string \"true\"", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: {
                  name: "data",
                  size: "1Gi",
                  attributes: { class: "ram", persistent: "true" },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"ram\" storage"),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("cannot be persistent"),
      }));
    });

    it("accepts RAM storage when not persistent", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: {
                  size: "1Gi",
                  attributes: { class: "ram" },
                },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
      });

      expect(validate()).toBeUndefined();
    });
  });

  describe("schema validation: pricing", () => {
    it("returns an error for invalid denom pattern", () => {
      const { validate } = setup({
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: "usdt" },
              },
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"denom\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("pattern \"^(uakt|ibc/.*)$\""),
      }));
    });

    it("returns an error when denom is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000" },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"denom\""),
      }));
    });

    it("returns an error when amount is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: { dcloud: { count: 1, profile: "web" } },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"amount\""),
      }));
    });
  });

  describe("schema validation: deployment", () => {
    it("returns an error when deployment count is 0", () => {
      const { validate } = setup({
        deployment: {
          web: {
            dcloud: {
              count: 0 as number,
              profile: "web",
            },
          },
        },
      });

      const errors = validate();

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"count\""),
      }));
      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("at least 1"),
      }));
    });

    it("returns an error when deployment profile is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: {
            dcloud: {
              count: 1,
            },
          },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"profile\""),
      }));
    });

    it("returns an error when deployment count is missing", () => {
      const sdl = {
        version: "2.0",
        services: {
          web: {
            image: "nginx:latest",
            expose: [{ port: 80, as: 80, to: [{ global: true }] }],
          },
        },
        profiles: {
          compute: {
            web: {
              resources: {
                cpu: { units: 1 },
                memory: { size: "512Mi" },
                storage: { size: "1Gi" },
              },
            },
          },
          placement: {
            dcloud: {
              pricing: {
                web: { amount: "1000", denom: AKT_DENOM },
              },
            },
          },
        },
        deployment: {
          web: {
            dcloud: {
              profile: "web",
            },
          },
        },
      } as unknown as SDLInput;

      const errors = validateSDL(sdl, "sandbox");

      expect(errors).toContainEqual(expect.objectContaining({
        message: expect.stringContaining("\"count\""),
      }));
    });
  });

  function setup(overrides: DeepPartial<SDLInput> = {}, networkId: NetworkId = "sandbox") {
    const defaultSDL: SDLInput = {
      version: "2.0",
      services: {
        web: {
          image: "nginx:latest",
          expose: [
            {
              port: 80,
              as: 80,
              to: [{ global: true }],
            },
          ],
        },
      },
      profiles: {
        compute: {
          web: {
            resources: {
              cpu: { units: 1 },
              memory: { size: "512Mi" },
              storage: { size: "1Gi" },
            },
          },
        },
        placement: {
          dcloud: {
            pricing: {
              web: {
                amount: "1000",
                denom: AKT_DENOM,
              },
            },
          },
        },
      },
      deployment: {
        web: {
          dcloud: {
            count: 1,
            profile: "web",
          },
        },
      },
    };

    const sdl = merge(defaultSDL, overrides);

    return {
      sdl,
      networkId,
      validate: () => validateSDL(sdl, networkId),
    };
  }
});

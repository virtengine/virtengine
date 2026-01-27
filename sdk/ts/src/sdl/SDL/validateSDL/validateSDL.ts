import { AKT_DENOM, type NetworkId, USDC_IBC_DENOMS } from "../../../network/index.ts";
import type { ErrorMessages, ValidationError, ValidationFunction } from "../../../utils/jsonSchemaValidation.ts";
import { dirname, getErrorLocation, humanizeErrors } from "../../../utils/jsonSchemaValidation.ts";
import { castArray, stringToBoolean } from "../utils.ts";
import { schema as validationSDLSchema, type SDLInput, validate as validateSDLInput } from "./validateSDLInput.ts";

export type { SDLInput };
export { validationSDLSchema };

const ERROR_MESSAGES: ErrorMessages = {
  "#/definitions/storageAttributesValidation"(error) {
    return `"ram" storage${getErrorLocation(dirname(error.instancePath))} cannot be persistent`;
  },
  "#/definitions/exposeToWithIpEnforcesGlobal/then/properties/global/const"() {
    return `If an IP is declared, the directive must be declared as global.`;
  },
};

export function validateSDL(sdl: SDLInput, networkId: NetworkId): undefined | ValidationError[] {
  validateSDLInput(sdl);
  const schemaErrors = humanizeErrors((validateSDLInput as ValidationFunction).errors, validationSDLSchema, ERROR_MESSAGES);
  if (schemaErrors.length) return schemaErrors;

  const validator = new SDLValidator(sdl);
  const errors = validator.validate(networkId);

  const allErrors = schemaErrors.concat(errors);
  return allErrors.length ? allErrors : undefined;
}

class SDLValidator {
  readonly #endpointsUsed = new Set<string>();
  readonly #portsUsed = new Map<string, string>();
  readonly #sdl: SDLInput;
  readonly #errors: ValidationError[] = [];

  constructor(sdl: SDLInput) {
    this.#sdl = sdl;
  }

  validate(networkId: NetworkId) {
    if (this.#sdl.services) {
      Object.keys(this.#sdl.services).forEach((serviceName) => {
        this.#validateDeploymentWithRelations(serviceName);
        this.#validateLeaseIP(serviceName);
      });
    }

    this.#validateDenom(networkId);
    this.#validateEndpoints();
    return this.#errors;
  }

  #validateDenom(networkId: NetworkId) {
    if (!this.#sdl.profiles?.placement) return;

    const usdcDenom = USDC_IBC_DENOMS[networkId];
    const denoms = Object.entries(this.#sdl.profiles.placement).map(([placementName, placement]) => {
      if (!placement.pricing) return [];
      return Object.entries(placement.pricing).map(([profile, pricing]) => ({
        path: `/profiles/placement/${placementName}/pricing/${profile}/denom`,
        denom: pricing.denom,
      }));
    }).flat();
    const invalidDenom = denoms.find(({ denom }) => denom !== AKT_DENOM && denom !== usdcDenom);
    if (invalidDenom) {
      this.#errors.push({
        message: `Invalid denom: "${invalidDenom.denom}" at path "${invalidDenom.path}". Only "uakt" and "${usdcDenom}" are supported.`,
        instancePath: invalidDenom.path,
        schemaPath: "#/definitions/priceCoin/properties/denom",
        keyword: "pattern",
        params: {
          pattern: "^(uakt|ibc/.*)$",
        },
      });
    }
  }

  #validateDeploymentWithRelations(serviceName: string) {
    const deployment = this.#sdl.deployment[serviceName];
    if (!deployment) {
      this.#errors.push({
        message: `Service "${serviceName}" is not defined at "/deployment" section.`,
        instancePath: `/deployment`,
        schemaPath: "#/properties/deployment",
        keyword: "required",
        params: {
          missingProperty: serviceName,
        },
      });
      return;
    }

    Object.keys(this.#sdl.deployment[serviceName]).forEach((deploymentName) => {
      this.#validateDeploymentRelations(serviceName, deploymentName);
      this.#validateServiceStorages(serviceName, deploymentName);
      this.#validateStorages(serviceName, deploymentName);
      this.#validateGPU(serviceName, deploymentName);
    });
  }

  #validateDeploymentRelations(serviceName: string, deploymentName: string) {
    const serviceDeployment = this.#sdl.deployment?.[serviceName]?.[deploymentName];
    const compute = this.#sdl.profiles?.compute?.[serviceDeployment?.profile];
    const infra = this.#sdl.profiles?.placement?.[deploymentName];

    if (!infra) {
      this.#errors.push({
        message: `The placement "${deploymentName}" is not defined in the "placement" section.`,
        instancePath: `/profiles/placement`,
        schemaPath: "#/properties/profiles/properties/placement",
        keyword: "required",
        params: {
          missingProperty: deploymentName,
        },
      });
    }

    if (infra && !infra.pricing?.[serviceDeployment?.profile]) {
      this.#errors.push({
        message: `The pricing for the "${serviceDeployment?.profile}" profile is not defined in the "${deploymentName}" placement.`,
        instancePath: `/profiles/placement/${deploymentName}/pricing`,
        schemaPath: "#/properties/profiles/properties/placement/additionalProperties/properties/pricing",
        keyword: "required",
        params: {
          missingProperty: serviceDeployment?.profile,
        },
      });
    }

    if (!compute) {
      this.#errors.push({
        message: `The compute requirements for the "${serviceDeployment?.profile}" profile are not defined in the "compute" section.`,
        instancePath: `/profiles/compute`,
        schemaPath: "#/properties/profiles/properties/compute",
        keyword: "required",
        params: {
          missingProperty: serviceDeployment?.profile,
        },
      });
    }
  }

  #validateServiceStorages(serviceName: string, deploymentName: string) {
    const service = this.#sdl.services?.[serviceName];
    const mounts: Record<string, string> = {};
    const serviceDeployment = this.#sdl.deployment[serviceName][deploymentName];
    const compute = this.#sdl.profiles?.compute?.[serviceDeployment.profile];
    const storages = castArray(compute?.resources.storage);

    if (!service?.params?.storage) {
      return;
    }

    Object.entries(service.params.storage).forEach(([storageName, storage]) => {
      if (!storage) {
        this.#errors.push({
          message: `Storage "${storageName}" is not configured.`,
          instancePath: `/services/${serviceName}/params/storage/${storageName}`,
          schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage/additionalProperties",
          keyword: "required",
          params: {
            missingProperty: storageName,
          },
        });
        return;
      }
      const storageNameExists = storages.some(({ name }) => name === storageName);
      if (!storageNameExists) {
        this.#errors.push({
          message: `Service "${serviceName}" references non-existing compute volume "${storageName}".`,
          instancePath: `/profiles/compute/${serviceDeployment.profile}/resources/storage`,
          schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/storage",
          keyword: "required",
          params: {
            missingProperty: storageName,
          },
        });
        return;
      }

      const mount = String(storage.mount);
      const volumeName = mounts[mount];

      if (volumeName && !storage.mount) {
        this.#errors.push({
          message: "Multiple root ephemeral storages are not allowed.",
          instancePath: `/services/${serviceName}/params/storage/${storageName}`,
          schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage",
          keyword: "uniqueItems",
          params: {
            duplicate: volumeName,
          },
        });
      }
      if (volumeName && storage.mount) {
        this.#errors.push({
          message: `Mount "${mount}" already in use by volume "${volumeName}".`,
          instancePath: `/services/${serviceName}/params/storage/${storageName}/mount`,
          schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage/additionalProperties/properties/mount",
          keyword: "uniqueItems",
          params: {
            duplicate: mount,
          },
        });
      }

      mounts[mount] = storageName;
    });
  }

  #validateStorages(serviceName: string, deploymentName: string) {
    const service = this.#sdl.services?.[serviceName];
    const serviceDeployment = this.#sdl.deployment[serviceName][deploymentName];
    const compute = this.#sdl.profiles?.compute?.[serviceDeployment.profile];
    const storages = castArray(compute?.resources.storage);

    storages.forEach((storage) => {
      const persistent = stringToBoolean(storage.attributes?.persistent as string | boolean || false);

      if (persistent && !service?.params?.storage?.[storage.name || ""]?.mount) {
        this.#errors.push({
          message: `Persistent storage "${storage.name || "default"}" requires a mount path in /services/${serviceName}/params/storage/${storage.name || "default"}/mount.`,
          instancePath: `/services/${serviceName}/params/storage/${storage.name || "default"}`,
          schemaPath: "#/properties/services/additionalProperties/properties/params/properties/storage/additionalProperties/properties/mount",
          keyword: "required",
          params: {
            missingProperty: "mount",
          },
        });
      }
    });
  }

  #validateGPU(serviceName: string, deploymentName: string) {
    const deployment = this.#sdl.deployment[serviceName];
    const compute = this.#sdl.profiles?.compute?.[deployment[deploymentName]?.profile];
    const gpu = compute?.resources.gpu;
    if (!gpu) return;

    const hasUnits = gpu.units !== undefined && gpu.units !== 0;
    const hasAttributes = typeof gpu.attributes !== "undefined";
    const hasVendor = hasAttributes && typeof gpu.attributes?.vendor !== "undefined";

    const profile = deployment[deploymentName]?.profile;
    const gpuPath = `/profiles/compute/${profile}/resources/gpu`;

    if (!hasUnits && hasAttributes) {
      this.#errors.push({
        message: "GPU must not have attributes if units is 0.",
        instancePath: `${gpuPath}/attributes`,
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/gpu/properties/attributes",
        keyword: "additionalProperties",
        params: {
          additionalProperty: "attributes",
        },
      });
    }
    if (hasUnits && !hasAttributes) {
      this.#errors.push({
        message: "GPU must have attributes if units is not 0.",
        instancePath: gpuPath,
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/gpu",
        keyword: "required",
        params: {
          missingProperty: "attributes",
        },
      });
    }
    if (hasUnits && !hasVendor) {
      this.#errors.push({
        message: "GPU must specify a vendor if units is not 0.",
        instancePath: `${gpuPath}/attributes`,
        schemaPath: "#/properties/profiles/properties/compute/additionalProperties/properties/resources/properties/gpu/properties/attributes/properties/vendor",
        keyword: "required",
        params: {
          missingProperty: "vendor",
        },
      });
    }
  }

  #validateLeaseIP(serviceName: string) {
    this.#sdl.services?.[serviceName]?.expose?.forEach((expose, exposeIndex) => {
      const proto = expose.proto?.toUpperCase() || "TCP";

      expose.to?.forEach((to, toIndex) => {
        if (to.ip?.length) {
          const toPath = `/services/${serviceName}/expose/${exposeIndex}/to/${toIndex}`;

          if (!to.global) {
            this.#errors.push({
              message: `If an IP is declared, the directive must be declared as global.`,
              instancePath: `${toPath}/global`,
              schemaPath: "#/definitions/exposeToWithIpEnforcesGlobal/then/properties/global/const",
              keyword: "const",
              params: {
                allowedValue: true,
              },
            });
          }
          if (!this.#sdl.endpoints?.[to.ip]) {
            this.#errors.push({
              message: `Unknown endpoint "${to.ip}" for service "${serviceName}". Add it to the "endpoints" section.`,
              instancePath: `/endpoints/${to.ip}`,
              schemaPath: "#/properties/endpoints",
              keyword: "required",
              params: {
                missingProperty: to.ip,
              },
            });
          }

          this.#endpointsUsed.add(to.ip);

          const externalPort = expose.as ?? expose.port;
          const portKey = `${to.ip}-${externalPort}-${proto}`;
          const otherServiceName = this.#portsUsed.get(portKey);

          if (this.#portsUsed.has(portKey)) {
            this.#errors.push({
              message: `IP endpoint "${to.ip}" port ${externalPort} protocol ${proto} already in use by service "${otherServiceName}".`,
              instancePath: `${toPath}/ip`,
              schemaPath: "#/properties/services/additionalProperties/properties/expose/items/properties/to/items",
              keyword: "uniqueItems",
              params: {
                duplicate: portKey,
              },
            });
          }
          this.#portsUsed.set(portKey, serviceName);
        }
      });
    });
  }

  #validateEndpoints() {
    if (!this.#sdl.endpoints) return;

    Object.keys(this.#sdl.endpoints).forEach((endpoint) => {
      if (!this.#endpointsUsed.has(endpoint)) {
        this.#errors.push({
          message: `Endpoint "${endpoint}" declared but never used.`,
          instancePath: `/endpoints/${endpoint}`,
          schemaPath: "#/properties/endpoints",
          keyword: "additionalProperties",
          params: {
            additionalProperty: endpoint,
          },
        });
      }
    });
  }
}

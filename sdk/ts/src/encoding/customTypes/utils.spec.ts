import type { DescField } from "@bufbuild/protobuf";
import { afterEach, describe, expect, it } from "@jest/globals";
import { proto } from "@test/helpers/proto";

import { LegacyDec } from "./LegacyDec.ts";
import { findPathsToCustomField } from "./utils.ts";

describe("utils", () => {
  describe(findPathsToCustomField.name, () => {
    afterEach(() => {
      findPathsToCustomField.clearCache();
    });

    it("returns empty array for the type which does not have custom types", async () => {
      const def = await proto`
        message Cat {
          string name = 1;
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("Cat"), () => true))).toEqual([]);
    });

    it("returns empty type for custom types that do not require patching", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message Dog {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          string name = 2;
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("Dog"), () => true))).toEqual([]);
    });

    it("returns paths to the fields which has custom types that require patching", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message MegaPuppy {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          string name = 2;
          string price = 3 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
          string age = 4 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("MegaPuppy"), () => true))).toEqual([
        "price",
        "age",
      ]);
    });

    it("returns nested message paths", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          Dog pet = 2;
        }

        message Dog {
          string name = 2;
          string price = 3 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetOwner"), () => true))).toEqual([
        "pet.price",
      ]);
    });

    it("returns message paths for \"oneof\" keyword", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          Pet pet = 2;
        }

        message Pet {
          oneof value {
            Dog dog = 2;
            Cat cat = 3;
          }
        }

        message Dog {
          string name = 2;
          string price = 3 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }

        message Cat {
          string name = 2;
          string price = 3;
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetOwner"), () => true))).toEqual([
        "pet.dog.price",
      ]);
    });

    it("returns message paths for repeated \"oneof\" keyword", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          repeated Pet pets = 2;
        }

        message Pet {
          oneof value {
            Dog dog = 2;
            Cat cat = 3;
          }
        }

        message Dog {
          string name = 2;
          string price = 3 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }

        message Cat {
          string name = 2;
          string price = 3;
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetOwner"), () => true))).toEqual([
        "pets.dog.price",
      ]);
    });

    it("returns nested repeated message paths", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          repeated Dog pets = 2;
        }

        message Dog {
          string name = 2;
          string price = 3 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetOwner"), () => true))).toEqual([
        "pets.price",
      ]);
    });

    it("returns paths for map values", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetClub {
          map<string, PetOwner> ownerToPet = 1;
        }

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          Pet pet = 2;
        }

        message Pet {
          string name = 2;
          string price = 3 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetClub"), () => true))).toEqual([
        "ownerToPet.pet.price",
      ]);
    });

    it("returns paths for cyclic refs", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message Attribute {
          Attribute parent = 1;
          string price = 2 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }

        message PetOwner {
          repeated Pet pets = 1;
          string cash = 2 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }

        message Pet {
          string price = 1 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
          PetOwner owner = 2;
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("Attribute"), () => true))).toEqual([
        "price",
      ]);
      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetOwner"), () => true))).toEqual([
        "pets.price",
        "cash",
      ]);
    });

    it("returns paths for complex combination of all cases", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          Pet pet = 2;
          Money cash = 3;
        }

        message Pet {
          oneof value {
            Dog dog = 2;
            Cat cat = 3;
          }
        }

        message PetClub {
          repeated PetOwner members = 1;
          repeated Dog all_dogs = 3;
          map<string, Dog> ownerIdToDog = 4;
        }

        message PetClubItemsQueryResponse {
          repeated PetClub items = 1;
        }

        message Dog {
          string name = 1 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }

        message Cat {
          string name = 1;
        }

        message Money {
          string currency = 1;
          string amount = 2 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }
      `;

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetClub"), () => true))).toEqual([
        "members.pet.dog.name",
        "members.cash.amount",
        "allDogs.name",
        "ownerIdToDog.name",
      ]);
    });

    it("returns paths only for messages that pass guard", async () => {
      const def = await proto`
        import "gogoproto/gogo.proto";

        message PetOwner {
          bytes id = 1 [
            (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int"
          ];
          Pet pet = 2;
          Money cash = 3;
        }

        message Pet {
          oneof value {
            Dog dog = 2;
            Cat cat = 3;
          }
        }

        message PetClub {
          repeated PetOwner members = 1;
          repeated Dog all_dogs = 3;
          map<string, Dog> ownerIdToDog = 4;
        }

        message PetClubItemsQueryResponse {
          repeated PetClub items = 1;
        }

        message Dog {
          string name = 1;
          string price = 2 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }

        message Cat {
          string name = 1;
        }

        message Money {
          string currency = 1;
          string amount = 2 [
            (gogoproto.customtype) = "${LegacyDec.typeName}"
          ];
        }
      `;
      const shouldCheckForCustomPaths = (typeName: string) => typeName.endsWith(".Money") || typeName.endsWith(".PetOwner");

      expect(stringifyPaths(findPathsToCustomField(def.getMessage("PetClub"), shouldCheckForCustomPaths))).toEqual([
        "members.cash.amount",
      ]);
    });

    function stringifyPaths(paths: DescField[][]): string[] {
      return paths.map((p) => p.map((f) => f.localName).join("."));
    }
  });
});

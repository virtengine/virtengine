import { LegacyDec } from "../../encoding/customTypes/LegacyDec.ts";
import type * as ..\protos\virtengine\bme\v1\types from "..\protos\virtengine\bme\v1\types.ts";
import type * as ..\protos\virtengine\bme\v1\events from "..\protos\virtengine\bme\v1\events.ts";
import type * as ..\protos\virtengine\bme\v1\genesis from "..\protos\virtengine\bme\v1\genesis.ts";
import type * as ..\protos\virtengine\bme\v1\query from "..\protos\virtengine\bme\v1\query.ts";
import type * as ..\protos\virtengine\deployment\v1beta3\resourceunit from "..\protos\virtengine\deployment\v1beta3\resourceunit.ts";
import type * as ..\protos\cosmos\base\v1beta1\coin from "..\protos\cosmos\base\v1beta1\coin.ts";
import type * as ..\protos\virtengine\deployment\v1beta3\groupspec from "..\protos\virtengine\deployment\v1beta3\groupspec.ts";
import type * as ..\protos\virtengine\deployment\v1beta3\deploymentmsg from "..\protos\virtengine\deployment\v1beta3\deploymentmsg.ts";
import type * as ..\protos\virtengine\deployment\v1beta3\group from "..\protos\virtengine\deployment\v1beta3\group.ts";
import type * as ..\protos\virtengine\deployment\v1beta3\genesis from "..\protos\virtengine\deployment\v1beta3\genesis.ts";
import type * as ..\protos\virtengine\deployment\v1beta4\resourceunit from "..\protos\virtengine\deployment\v1beta4\resourceunit.ts";
import type * as ..\protos\virtengine\deployment\v1beta4\groupspec from "..\protos\virtengine\deployment\v1beta4\groupspec.ts";
import type * as ..\protos\virtengine\deployment\v1beta4\deploymentmsg from "..\protos\virtengine\deployment\v1beta4\deploymentmsg.ts";
import type * as ..\protos\virtengine\deployment\v1beta4\group from "..\protos\virtengine\deployment\v1beta4\group.ts";
import type * as ..\protos\virtengine\deployment\v1beta4\genesis from "..\protos\virtengine\deployment\v1beta4\genesis.ts";
import type * as ..\protos\virtengine\escrow\types\v1\balance from "..\protos\virtengine\escrow\types\v1\balance.ts";
import type * as ..\protos\virtengine\escrow\types\v1\deposit from "..\protos\virtengine\escrow\types\v1\deposit.ts";
import type * as ..\protos\virtengine\escrow\types\v1\account from "..\protos\virtengine\escrow\types\v1\account.ts";
import type * as ..\protos\virtengine\deployment\v1beta4\query from "..\protos\virtengine\deployment\v1beta4\query.ts";
import type * as ..\protos\virtengine\deployment\v1beta5\resourceunit from "..\protos\virtengine\deployment\v1beta5\resourceunit.ts";
import type * as ..\protos\virtengine\deployment\v1beta5\groupspec from "..\protos\virtengine\deployment\v1beta5\groupspec.ts";
import type * as ..\protos\virtengine\deployment\v1beta5\deploymentmsg from "..\protos\virtengine\deployment\v1beta5\deploymentmsg.ts";
import type * as ..\protos\virtengine\deployment\v1beta5\group from "..\protos\virtengine\deployment\v1beta5\group.ts";
import type * as ..\protos\virtengine\deployment\v1beta5\genesis from "..\protos\virtengine\deployment\v1beta5\genesis.ts";
import type * as ..\protos\virtengine\deployment\v1beta5\query from "..\protos\virtengine\deployment\v1beta5\query.ts";
import type * as ..\protos\virtengine\escrow\types\v1\payment from "..\protos\virtengine\escrow\types\v1\payment.ts";
import type * as ..\protos\virtengine\escrow\v1\genesis from "..\protos\virtengine\escrow\v1\genesis.ts";
import type * as ..\protos\virtengine\escrow\v1\query from "..\protos\virtengine\escrow\v1\query.ts";
import type * as ..\protos\virtengine\escrow\v1beta3\types from "..\protos\virtengine\escrow\v1beta3\types.ts";
import type * as ..\protos\virtengine\escrow\v1beta3\genesis from "..\protos\virtengine\escrow\v1beta3\genesis.ts";
import type * as ..\protos\virtengine\escrow\v1beta3\query from "..\protos\virtengine\escrow\v1beta3\query.ts";
import type * as ..\protos\virtengine\market\v1\lease from "..\protos\virtengine\market\v1\lease.ts";
import type * as ..\protos\virtengine\market\v1\event from "..\protos\virtengine\market\v1\event.ts";
import type * as ..\protos\virtengine\market\v1beta4\order from "..\protos\virtengine\market\v1beta4\order.ts";
import type * as ..\protos\virtengine\market\v1beta4\bid from "..\protos\virtengine\market\v1beta4\bid.ts";
import type * as ..\protos\virtengine\market\v1beta4\lease from "..\protos\virtengine\market\v1beta4\lease.ts";
import type * as ..\protos\virtengine\market\v1beta4\genesis from "..\protos\virtengine\market\v1beta4\genesis.ts";
import type * as ..\protos\virtengine\market\v1beta5\bid from "..\protos\virtengine\market\v1beta5\bid.ts";
import type * as ..\protos\virtengine\market\v1beta5\bidmsg from "..\protos\virtengine\market\v1beta5\bidmsg.ts";
import type * as ..\protos\virtengine\market\v1beta5\order from "..\protos\virtengine\market\v1beta5\order.ts";
import type * as ..\protos\virtengine\market\v1beta5\genesis from "..\protos\virtengine\market\v1beta5\genesis.ts";
import type * as ..\protos\virtengine\market\v1beta5\query from "..\protos\virtengine\market\v1beta5\query.ts";
import type * as ..\protos\virtengine\market\v2beta1\bid from "..\protos\virtengine\market\v2beta1\bid.ts";
import type * as ..\protos\virtengine\market\v2beta1\bidmsg from "..\protos\virtengine\market\v2beta1\bidmsg.ts";
import type * as ..\protos\virtengine\market\v2beta1\order from "..\protos\virtengine\market\v2beta1\order.ts";
import type * as ..\protos\virtengine\market\v2beta1\lease from "..\protos\virtengine\market\v2beta1\lease.ts";
import type * as ..\protos\virtengine\market\v2beta1\event from "..\protos\virtengine\market\v2beta1\event.ts";
import type * as ..\protos\virtengine\market\v2beta1\genesis from "..\protos\virtengine\market\v2beta1\genesis.ts";
import type * as ..\protos\virtengine\market\v2beta1\query from "..\protos\virtengine\market\v2beta1\query.ts";
import type * as ..\protos\virtengine\oracle\v1\prices from "..\protos\virtengine\oracle\v1\prices.ts";
import type * as ..\protos\virtengine\oracle\v1\events from "..\protos\virtengine\oracle\v1\events.ts";
import type * as ..\protos\virtengine\oracle\v1\genesis from "..\protos\virtengine\oracle\v1\genesis.ts";
import type * as ..\protos\virtengine\oracle\v1\msgs from "..\protos\virtengine\oracle\v1\msgs.ts";
import type * as ..\protos\virtengine\oracle\v1\query from "..\protos\virtengine\oracle\v1\query.ts";

const p = {
  "virtengine.bme.v1.CollateralRatio"(value: ..\protos\virtengine\bme\v1\types.CollateralRatio | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.ratio != null) newValue.ratio = LegacyDec[transformType](value.ratio);
    if (value.referencePrice != null) newValue.referencePrice = LegacyDec[transformType](value.referencePrice);
    return newValue;
  },
  "virtengine.bme.v1.CoinPrice"(value: ..\protos\virtengine\bme\v1\types.CoinPrice | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = LegacyDec[transformType](value.price);
    return newValue;
  },
  "virtengine.bme.v1.BurnMintPair"(value: ..\protos\virtengine\bme\v1\types.BurnMintPair | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.burned != null) newValue.burned = p["virtengine.bme.v1.CoinPrice"](value.burned, transformType);
    if (value.minted != null) newValue.minted = p["virtengine.bme.v1.CoinPrice"](value.minted, transformType);
    return newValue;
  },
  "virtengine.bme.v1.LedgerRecord"(value: ..\protos\virtengine\bme\v1\types.LedgerRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.burned != null) newValue.burned = p["virtengine.bme.v1.CoinPrice"](value.burned, transformType);
    if (value.minted != null) newValue.minted = p["virtengine.bme.v1.CoinPrice"](value.minted, transformType);
    if (value.remintCreditIssued != null) newValue.remintCreditIssued = p["virtengine.bme.v1.CoinPrice"](value.remintCreditIssued, transformType);
    if (value.remintCreditAccrued != null) newValue.remintCreditAccrued = p["virtengine.bme.v1.CoinPrice"](value.remintCreditAccrued, transformType);
    return newValue;
  },
  "virtengine.bme.v1.EventMintStatusChange"(value: ..\protos\virtengine\bme\v1\events.EventMintStatusChange | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.collateralRatio != null) newValue.collateralRatio = LegacyDec[transformType](value.collateralRatio);
    return newValue;
  },
  "virtengine.bme.v1.GenesisLedgerRecord"(value: ..\protos\virtengine\bme\v1\genesis.GenesisLedgerRecord | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.record != null) newValue.record = p["virtengine.bme.v1.LedgerRecord"](value.record, transformType);
    return newValue;
  },
  "virtengine.bme.v1.GenesisLedgerState"(value: ..\protos\virtengine\bme\v1\genesis.GenesisLedgerState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.records) newValue.records = value.records.map((item) => p["virtengine.bme.v1.GenesisLedgerRecord"](item, transformType)!);
    return newValue;
  },
  "virtengine.bme.v1.GenesisState"(value: ..\protos\virtengine\bme\v1\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.ledger != null) newValue.ledger = p["virtengine.bme.v1.GenesisLedgerState"](value.ledger, transformType);
    return newValue;
  },
  "virtengine.bme.v1.QueryStatusResponse"(value: ..\protos\virtengine\bme\v1\query.QueryStatusResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.collateralRatio != null) newValue.collateralRatio = LegacyDec[transformType](value.collateralRatio);
    if (value.warnThreshold != null) newValue.warnThreshold = LegacyDec[transformType](value.warnThreshold);
    if (value.haltThreshold != null) newValue.haltThreshold = LegacyDec[transformType](value.haltThreshold);
    return newValue;
  },
  "virtengine.deployment.v1beta3.ResourceUnit"(value: ..\protos\virtengine\deployment\v1beta3\resourceunit.ResourceUnit | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "cosmos.base.v1beta1.DecCoin"(value: ..\protos\cosmos\base\v1beta1\coin.DecCoin | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.amount != null) newValue.amount = LegacyDec[transformType](value.amount);
    return newValue;
  },
  "virtengine.deployment.v1beta3.GroupSpec"(value: ..\protos\virtengine\deployment\v1beta3\groupspec.GroupSpec | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.resources) newValue.resources = value.resources.map((item) => p["virtengine.deployment.v1beta3.ResourceUnit"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta3.MsgCreateDeployment"(value: ..\protos\virtengine\deployment\v1beta3\deploymentmsg.MsgCreateDeployment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta3.GroupSpec"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta3.Group"(value: ..\protos\virtengine\deployment\v1beta3\group.Group | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groupSpec != null) newValue.groupSpec = p["virtengine.deployment.v1beta3.GroupSpec"](value.groupSpec, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta3.GenesisDeployment"(value: ..\protos\virtengine\deployment\v1beta3\genesis.GenesisDeployment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta3.Group"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta3.GenesisState"(value: ..\protos\virtengine\deployment\v1beta3\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.deployments) newValue.deployments = value.deployments.map((item) => p["virtengine.deployment.v1beta3.GenesisDeployment"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta4.ResourceUnit"(value: ..\protos\virtengine\deployment\v1beta4\resourceunit.ResourceUnit | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta4.GroupSpec"(value: ..\protos\virtengine\deployment\v1beta4\groupspec.GroupSpec | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.resources) newValue.resources = value.resources.map((item) => p["virtengine.deployment.v1beta4.ResourceUnit"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta4.MsgCreateDeployment"(value: ..\protos\virtengine\deployment\v1beta4\deploymentmsg.MsgCreateDeployment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta4.GroupSpec"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta4.Group"(value: ..\protos\virtengine\deployment\v1beta4\group.Group | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groupSpec != null) newValue.groupSpec = p["virtengine.deployment.v1beta4.GroupSpec"](value.groupSpec, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta4.GenesisDeployment"(value: ..\protos\virtengine\deployment\v1beta4\genesis.GenesisDeployment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta4.Group"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta4.GenesisState"(value: ..\protos\virtengine\deployment\v1beta4\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.deployments) newValue.deployments = value.deployments.map((item) => p["virtengine.deployment.v1beta4.GenesisDeployment"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.types.v1.Balance"(value: ..\protos\virtengine\escrow\types\v1\balance.Balance | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.amount != null) newValue.amount = LegacyDec[transformType](value.amount);
    return newValue;
  },
  "virtengine.escrow.types.v1.Depositor"(value: ..\protos\virtengine\escrow\types\v1\deposit.Depositor | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.balance != null) newValue.balance = p["cosmos.base.v1beta1.DecCoin"](value.balance, transformType);
    return newValue;
  },
  "virtengine.escrow.types.v1.AccountState"(value: ..\protos\virtengine\escrow\types\v1\account.AccountState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.transferred) newValue.transferred = value.transferred.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    if (value.funds) newValue.funds = value.funds.map((item) => p["virtengine.escrow.types.v1.Balance"](item, transformType)!);
    if (value.deposits) newValue.deposits = value.deposits.map((item) => p["virtengine.escrow.types.v1.Depositor"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.types.v1.Account"(value: ..\protos\virtengine\escrow\types\v1\account.Account | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.state != null) newValue.state = p["virtengine.escrow.types.v1.AccountState"](value.state, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta4.QueryDeploymentsResponse"(value: ..\protos\virtengine\deployment\v1beta4\query.QueryDeploymentsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.deployments) newValue.deployments = value.deployments.map((item) => p["virtengine.deployment.v1beta4.QueryDeploymentResponse"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta4.QueryDeploymentResponse"(value: ..\protos\virtengine\deployment\v1beta4\query.QueryDeploymentResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta4.Group"](item, transformType)!);
    if (value.escrowAccount != null) newValue.escrowAccount = p["virtengine.escrow.types.v1.Account"](value.escrowAccount, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta4.QueryGroupResponse"(value: ..\protos\virtengine\deployment\v1beta4\query.QueryGroupResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.group != null) newValue.group = p["virtengine.deployment.v1beta4.Group"](value.group, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta5.ResourceUnit"(value: ..\protos\virtengine\deployment\v1beta5\resourceunit.ResourceUnit | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta5.GroupSpec"(value: ..\protos\virtengine\deployment\v1beta5\groupspec.GroupSpec | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.resources) newValue.resources = value.resources.map((item) => p["virtengine.deployment.v1beta5.ResourceUnit"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta5.MsgCreateDeployment"(value: ..\protos\virtengine\deployment\v1beta5\deploymentmsg.MsgCreateDeployment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta5.GroupSpec"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta5.Group"(value: ..\protos\virtengine\deployment\v1beta5\group.Group | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groupSpec != null) newValue.groupSpec = p["virtengine.deployment.v1beta5.GroupSpec"](value.groupSpec, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta5.GenesisDeployment"(value: ..\protos\virtengine\deployment\v1beta5\genesis.GenesisDeployment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta5.Group"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta5.GenesisState"(value: ..\protos\virtengine\deployment\v1beta5\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.deployments) newValue.deployments = value.deployments.map((item) => p["virtengine.deployment.v1beta5.GenesisDeployment"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta5.QueryDeploymentsResponse"(value: ..\protos\virtengine\deployment\v1beta5\query.QueryDeploymentsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.deployments) newValue.deployments = value.deployments.map((item) => p["virtengine.deployment.v1beta5.QueryDeploymentResponse"](item, transformType)!);
    return newValue;
  },
  "virtengine.deployment.v1beta5.QueryDeploymentResponse"(value: ..\protos\virtengine\deployment\v1beta5\query.QueryDeploymentResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.groups) newValue.groups = value.groups.map((item) => p["virtengine.deployment.v1beta5.Group"](item, transformType)!);
    if (value.escrowAccount != null) newValue.escrowAccount = p["virtengine.escrow.types.v1.Account"](value.escrowAccount, transformType);
    return newValue;
  },
  "virtengine.deployment.v1beta5.QueryGroupResponse"(value: ..\protos\virtengine\deployment\v1beta5\query.QueryGroupResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.group != null) newValue.group = p["virtengine.deployment.v1beta5.Group"](value.group, transformType);
    return newValue;
  },
  "virtengine.escrow.types.v1.PaymentState"(value: ..\protos\virtengine\escrow\types\v1\payment.PaymentState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rate != null) newValue.rate = p["cosmos.base.v1beta1.DecCoin"](value.rate, transformType);
    if (value.balance != null) newValue.balance = p["cosmos.base.v1beta1.DecCoin"](value.balance, transformType);
    if (value.unsettled != null) newValue.unsettled = p["cosmos.base.v1beta1.DecCoin"](value.unsettled, transformType);
    return newValue;
  },
  "virtengine.escrow.types.v1.Payment"(value: ..\protos\virtengine\escrow\types\v1\payment.Payment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.state != null) newValue.state = p["virtengine.escrow.types.v1.PaymentState"](value.state, transformType);
    return newValue;
  },
  "virtengine.escrow.v1.GenesisState"(value: ..\protos\virtengine\escrow\v1\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.accounts) newValue.accounts = value.accounts.map((item) => p["virtengine.escrow.types.v1.Account"](item, transformType)!);
    if (value.payments) newValue.payments = value.payments.map((item) => p["virtengine.escrow.types.v1.Payment"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.v1.QueryAccountsResponse"(value: ..\protos\virtengine\escrow\v1\query.QueryAccountsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.accounts) newValue.accounts = value.accounts.map((item) => p["virtengine.escrow.types.v1.Account"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.v1.QueryPaymentsResponse"(value: ..\protos\virtengine\escrow\v1\query.QueryPaymentsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.payments) newValue.payments = value.payments.map((item) => p["virtengine.escrow.types.v1.Payment"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.v1beta3.Account"(value: ..\protos\virtengine\escrow\v1beta3\types.Account | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.balance != null) newValue.balance = p["cosmos.base.v1beta1.DecCoin"](value.balance, transformType);
    if (value.transferred != null) newValue.transferred = p["cosmos.base.v1beta1.DecCoin"](value.transferred, transformType);
    if (value.funds != null) newValue.funds = p["cosmos.base.v1beta1.DecCoin"](value.funds, transformType);
    return newValue;
  },
  "virtengine.escrow.v1beta3.FractionalPayment"(value: ..\protos\virtengine\escrow\v1beta3\types.FractionalPayment | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.rate != null) newValue.rate = p["cosmos.base.v1beta1.DecCoin"](value.rate, transformType);
    if (value.balance != null) newValue.balance = p["cosmos.base.v1beta1.DecCoin"](value.balance, transformType);
    return newValue;
  },
  "virtengine.escrow.v1beta3.GenesisState"(value: ..\protos\virtengine\escrow\v1beta3\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.accounts) newValue.accounts = value.accounts.map((item) => p["virtengine.escrow.v1beta3.Account"](item, transformType)!);
    if (value.payments) newValue.payments = value.payments.map((item) => p["virtengine.escrow.v1beta3.FractionalPayment"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.v1beta3.QueryAccountsResponse"(value: ..\protos\virtengine\escrow\v1beta3\query.QueryAccountsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.accounts) newValue.accounts = value.accounts.map((item) => p["virtengine.escrow.v1beta3.Account"](item, transformType)!);
    return newValue;
  },
  "virtengine.escrow.v1beta3.QueryPaymentsResponse"(value: ..\protos\virtengine\escrow\v1beta3\query.QueryPaymentsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.payments) newValue.payments = value.payments.map((item) => p["virtengine.escrow.v1beta3.FractionalPayment"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v1.Lease"(value: ..\protos\virtengine\market\v1\lease.Lease | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1.EventBidCreated"(value: ..\protos\virtengine\market\v1\event.EventBidCreated | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1.EventLeaseCreated"(value: ..\protos\virtengine\market\v1\event.EventLeaseCreated | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1beta4.Order"(value: ..\protos\virtengine\market\v1beta4\order.Order | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.spec != null) newValue.spec = p["virtengine.deployment.v1beta3.GroupSpec"](value.spec, transformType);
    return newValue;
  },
  "virtengine.market.v1beta4.MsgCreateBid"(value: ..\protos\virtengine\market\v1beta4\bid.MsgCreateBid | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1beta4.Bid"(value: ..\protos\virtengine\market\v1beta4\bid.Bid | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1beta4.Lease"(value: ..\protos\virtengine\market\v1beta4\lease.Lease | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1beta4.GenesisState"(value: ..\protos\virtengine\market\v1beta4\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.orders) newValue.orders = value.orders.map((item) => p["virtengine.market.v1beta4.Order"](item, transformType)!);
    if (value.leases) newValue.leases = value.leases.map((item) => p["virtengine.market.v1beta4.Lease"](item, transformType)!);
    if (value.bids) newValue.bids = value.bids.map((item) => p["virtengine.market.v1beta4.Bid"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v1beta5.Bid"(value: ..\protos\virtengine\market\v1beta5\bid.Bid | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1beta5.MsgCreateBid"(value: ..\protos\virtengine\market\v1beta5\bidmsg.MsgCreateBid | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["cosmos.base.v1beta1.DecCoin"](value.price, transformType);
    return newValue;
  },
  "virtengine.market.v1beta5.Order"(value: ..\protos\virtengine\market\v1beta5\order.Order | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.spec != null) newValue.spec = p["virtengine.deployment.v1beta4.GroupSpec"](value.spec, transformType);
    return newValue;
  },
  "virtengine.market.v1beta5.GenesisState"(value: ..\protos\virtengine\market\v1beta5\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.orders) newValue.orders = value.orders.map((item) => p["virtengine.market.v1beta5.Order"](item, transformType)!);
    if (value.leases) newValue.leases = value.leases.map((item) => p["virtengine.market.v1.Lease"](item, transformType)!);
    if (value.bids) newValue.bids = value.bids.map((item) => p["virtengine.market.v1beta5.Bid"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v1beta5.QueryOrdersResponse"(value: ..\protos\virtengine\market\v1beta5\query.QueryOrdersResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.orders) newValue.orders = value.orders.map((item) => p["virtengine.market.v1beta5.Order"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v1beta5.QueryOrderResponse"(value: ..\protos\virtengine\market\v1beta5\query.QueryOrderResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.order != null) newValue.order = p["virtengine.market.v1beta5.Order"](value.order, transformType);
    return newValue;
  },
  "virtengine.market.v1beta5.QueryBidsResponse"(value: ..\protos\virtengine\market\v1beta5\query.QueryBidsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.bids) newValue.bids = value.bids.map((item) => p["virtengine.market.v1beta5.QueryBidResponse"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v1beta5.QueryBidResponse"(value: ..\protos\virtengine\market\v1beta5\query.QueryBidResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.bid != null) newValue.bid = p["virtengine.market.v1beta5.Bid"](value.bid, transformType);
    if (value.escrowAccount != null) newValue.escrowAccount = p["virtengine.escrow.types.v1.Account"](value.escrowAccount, transformType);
    return newValue;
  },
  "virtengine.market.v1beta5.QueryLeasesResponse"(value: ..\protos\virtengine\market\v1beta5\query.QueryLeasesResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.leases) newValue.leases = value.leases.map((item) => p["virtengine.market.v1beta5.QueryLeaseResponse"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v1beta5.QueryLeaseResponse"(value: ..\protos\virtengine\market\v1beta5\query.QueryLeaseResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.lease != null) newValue.lease = p["virtengine.market.v1.Lease"](value.lease, transformType);
    if (value.escrowPayment != null) newValue.escrowPayment = p["virtengine.escrow.types.v1.Payment"](value.escrowPayment, transformType);
    return newValue;
  },
  "virtengine.market.v2beta1.Bid"(value: ..\protos\virtengine\market\v2beta1\bid.Bid | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.MsgCreateBid"(value: ..\protos\virtengine\market\v2beta1\bidmsg.MsgCreateBid | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.Order"(value: ..\protos\virtengine\market\v2beta1\order.Order | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.spec != null) newValue.spec = p["virtengine.deployment.v1beta5.GroupSpec"](value.spec, transformType);
    return newValue;
  },
  "virtengine.market.v2beta1.Lease"(value: ..\protos\virtengine\market\v2beta1\lease.Lease | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.EventBidCreated"(value: ..\protos\virtengine\market\v2beta1\event.EventBidCreated | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.EventLeaseCreated"(value: ..\protos\virtengine\market\v2beta1\event.EventLeaseCreated | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["cosmos.base.v1beta1.DecCoin"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.GenesisState"(value: ..\protos\virtengine\market\v2beta1\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.orders) newValue.orders = value.orders.map((item) => p["virtengine.market.v2beta1.Order"](item, transformType)!);
    if (value.leases) newValue.leases = value.leases.map((item) => p["virtengine.market.v2beta1.Lease"](item, transformType)!);
    if (value.bids) newValue.bids = value.bids.map((item) => p["virtengine.market.v2beta1.Bid"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.QueryOrdersResponse"(value: ..\protos\virtengine\market\v2beta1\query.QueryOrdersResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.orders) newValue.orders = value.orders.map((item) => p["virtengine.market.v2beta1.Order"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.QueryOrderResponse"(value: ..\protos\virtengine\market\v2beta1\query.QueryOrderResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.order != null) newValue.order = p["virtengine.market.v2beta1.Order"](value.order, transformType);
    return newValue;
  },
  "virtengine.market.v2beta1.QueryBidsResponse"(value: ..\protos\virtengine\market\v2beta1\query.QueryBidsResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.bids) newValue.bids = value.bids.map((item) => p["virtengine.market.v2beta1.QueryBidResponse"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.QueryBidResponse"(value: ..\protos\virtengine\market\v2beta1\query.QueryBidResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.bid != null) newValue.bid = p["virtengine.market.v2beta1.Bid"](value.bid, transformType);
    if (value.escrowAccount != null) newValue.escrowAccount = p["virtengine.escrow.types.v1.Account"](value.escrowAccount, transformType);
    return newValue;
  },
  "virtengine.market.v2beta1.QueryLeasesResponse"(value: ..\protos\virtengine\market\v2beta1\query.QueryLeasesResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.leases) newValue.leases = value.leases.map((item) => p["virtengine.market.v2beta1.QueryLeaseResponse"](item, transformType)!);
    return newValue;
  },
  "virtengine.market.v2beta1.QueryLeaseResponse"(value: ..\protos\virtengine\market\v2beta1\query.QueryLeaseResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.lease != null) newValue.lease = p["virtengine.market.v2beta1.Lease"](value.lease, transformType);
    if (value.escrowPayment != null) newValue.escrowPayment = p["virtengine.escrow.types.v1.Payment"](value.escrowPayment, transformType);
    return newValue;
  },
  "virtengine.oracle.v1.PriceDataState"(value: ..\protos\virtengine\oracle\v1\prices.PriceDataState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = LegacyDec[transformType](value.price);
    return newValue;
  },
  "virtengine.oracle.v1.PriceData"(value: ..\protos\virtengine\oracle\v1\prices.PriceData | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.state != null) newValue.state = p["virtengine.oracle.v1.PriceDataState"](value.state, transformType);
    return newValue;
  },
  "virtengine.oracle.v1.AggregatedPrice"(value: ..\protos\virtengine\oracle\v1\prices.AggregatedPrice | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.twap != null) newValue.twap = LegacyDec[transformType](value.twap);
    if (value.medianPrice != null) newValue.medianPrice = LegacyDec[transformType](value.medianPrice);
    if (value.minPrice != null) newValue.minPrice = LegacyDec[transformType](value.minPrice);
    if (value.maxPrice != null) newValue.maxPrice = LegacyDec[transformType](value.maxPrice);
    return newValue;
  },
  "virtengine.oracle.v1.QueryPricesResponse"(value: ..\protos\virtengine\oracle\v1\prices.QueryPricesResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["virtengine.oracle.v1.PriceData"](item, transformType)!);
    return newValue;
  },
  "virtengine.oracle.v1.EventPriceData"(value: ..\protos\virtengine\oracle\v1\events.EventPriceData | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.data != null) newValue.data = p["virtengine.oracle.v1.PriceDataState"](value.data, transformType);
    return newValue;
  },
  "virtengine.oracle.v1.GenesisState"(value: ..\protos\virtengine\oracle\v1\genesis.GenesisState | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.prices) newValue.prices = value.prices.map((item) => p["virtengine.oracle.v1.PriceData"](item, transformType)!);
    return newValue;
  },
  "virtengine.oracle.v1.MsgAddPriceEntry"(value: ..\protos\virtengine\oracle\v1\msgs.MsgAddPriceEntry | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.price != null) newValue.price = p["virtengine.oracle.v1.PriceDataState"](value.price, transformType);
    return newValue;
  },
  "virtengine.oracle.v1.QueryAggregatedPriceResponse"(value: ..\protos\virtengine\oracle\v1\query.QueryAggregatedPriceResponse | undefined | null, transformType: 'encode' | 'decode') {
    if (value == null) return;
    const newValue = { ...value };
    if (value.aggregatedPrice != null) newValue.aggregatedPrice = p["virtengine.oracle.v1.AggregatedPrice"](value.aggregatedPrice, transformType);
    return newValue;
  }
};

export const patches = p;

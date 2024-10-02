import localforage from "localforage";
import { matchSorter } from "match-sorter";
import { sortBy } from 'sort-by-typescript';
import { ChannelSettings } from "../models/models";


export async function getChannelSettings(query?:string) {
  await fakeNetwork(`getChannelSettings:${query}`);
  let allSettings : ChannelSettings[] | null = await localforage.getItem("settings");
  if (!allSettings) allSettings = [];
  if (query) {
    allSettings = matchSorter(allSettings, query, { keys: ["channelName"] });
  }
  return allSettings.sort(sortBy("channelName"));
}

export async function createUserChannelSettings(channelName: string) {
  await fakeNetwork();
  // let id = Math.random().toString(36).substring(2, 9);
  let id = channelName;
  let userSetting :ChannelSettings = {
    channelName: id,
    minAnimationSpeed: 0,
    maxAnimationSpeed: 10,
    minVelocity: 0,
    maxVelocity: 10,
    minSpriteScale: 0,
    maxSpriteScale: 10,
    maxSpritePixelSize: 1000
  }
  let allSettings = await getChannelSettings();
  allSettings.unshift(userSetting);
  await set(allSettings);
  return userSetting;
}

export async function getUserChannelSettings(channelName:string) {
  await fakeNetwork(`channelSetting:${channelName}`);
  let allSettings: ChannelSettings[]|null = await localforage.getItem("settings");
  let setting = allSettings?.find(setting => setting.channelName === channelName);
  return setting ?? null;
}

export async function updateUserChannelSettings(channelName:string, updates:ChannelSettings) {
  await fakeNetwork();
  let allSettings: ChannelSettings[]|null = await localforage.getItem("settings");
  let setting = allSettings?.find(setting => setting.channelName === channelName);
  if (!setting) throw new Error("No setting found for " + channelName);
  Object.assign(setting, updates);
  await set(allSettings!);
  return setting;
}

export async function deleteUserChannelSettings(channelName:string) {
  let allSettings: ChannelSettings[] | null = await localforage.getItem("settings")!;
  let index = allSettings!.findIndex(contact => contact.channelName === channelName);
  if (index > -1) {
    allSettings!.splice(index, 1);
    await set(allSettings!);
    return true;
  }
  return false;
}

function set(settings: ChannelSettings[]) {
  return localforage.setItem("settings", settings);
}

// fake a cache so we don't slow down stuff we've already seen
let fakeCache: any = {};

async function fakeNetwork(key ?: any) {
  if (!key) {
    fakeCache = {};
  }

  if (fakeCache[key]) {
    return;
  }

  fakeCache[key] = true;
  return new Promise(res => {
    setTimeout(res, Math.random() * 800);
  });
}

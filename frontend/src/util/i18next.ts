import { writable } from 'svelte/store';
import i18next, { type TFunction } from 'i18next';
import Fetch from 'i18next-fetch-backend';


let ttt = (async () => {
  const t = await i18next
    .use(Fetch)
    .init({
      ns: 'skiver',
      lng: 'en',
      preload: ['en', 'nb'],
      saveMissing: true,
      // resources,
      debug: false,
      backend: {
        loadPath: "/api/export/l={{lng}}&p={{ns}}.json",
        addPath: '/api/missing/{{lng}}/{{ns}}',
      },

      fallbackLng: 'en',
      defaultNS: 'skiver',
      fallbackNS: 'skiver',
    });

  await new Promise(res => i18next.loadResources(res));

  return t
})()
function createL10nStore() {
  const { subscribe, update } = writable<TFunction>(() => "...initializing...");

  async function init() {
    let t = await ttt
    update(_t => t);
  }

  async function changeLanguage(lang: string) {
    let t = await i18next.changeLanguage(lang);
    update(_t => t);
  }

  return {
    subscribe,
    init,
    changeLanguage,
    i18next,
  };
}

export const t = createL10nStore();
t.init()

/** Used to add translation-resources for use in preview.*/
export function addPreviewTranslationResource(localeKey: string, ns: string, categoryKey: string, key: string, value: string, context?: string) {
  if (!localeKey) {
    return
  }
  if (!ns) {
    return
  }
  if (!categoryKey) {
    return
  }
  if (!key) {
    return
  }
  const cat = categoryKey === '___root___' ? "" : categoryKey
  const kk = context ? key + "_" + context : key
  const k = cat ? cat + "." + kk : kk
  // TODO: check if this is performance-heavy.
  i18next.addResource(localeKey, '__derived__' + ns, k, value)
}


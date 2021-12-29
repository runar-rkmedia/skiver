import createStore from './store'

export const state = createStore({
  initialValue: {
    showDeleted: false,
    serverStats: false,
    seenHints: {} as Record<string, [version: number, readAt: Date]>,
    collapse: {} as Record<string, boolean>,
    createTranslation: {} as ApiDef.TranslationInput,
  },
  storage: {
    key: 'state',
  },
})

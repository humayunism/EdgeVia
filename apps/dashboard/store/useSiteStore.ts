import { create } from 'zustand'

interface Site {
  id: string
  domain: string
  originUrl: string
  rpsLimit: number
  active: boolean
}

interface SiteStore {
  sites: Site[]
  activeSiteId: string | null
  setSites: (sites: Site[]) => void
  setActiveSite: (id: string) => void
  updateSiteConfig: (id: string, config: Partial<Site>) => void
}

export const useSiteStore = create<SiteStore>((set) => ({
  sites: [],
  activeSiteId: null,
  setSites: (sites) => set({ sites, activeSiteId: sites[0]?.id ?? null }),
  setActiveSite: (id) => set({ activeSiteId: id }),
  updateSiteConfig: (id, config) =>
    set((state) => ({
      sites: state.sites.map((s) => (s.id === id ? { ...s, ...config } : s))
    }))
}))

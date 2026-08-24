<template>
  <div class="home-showcase" :class="{ 'is-dark': isDark }" data-testid="home-showcase">
    <div class="showcase-grid" aria-hidden="true"></div>

    <header class="showcase-header">
      <nav class="showcase-nav" aria-label="Primary navigation">
        <router-link to="/home" class="showcase-brand" aria-label="Home">
          <span class="brand-mark"><img :src="siteLogo || '/logo.svg'" :alt="siteName" /></span>
          <span class="brand-copy">
            <strong>{{ siteName }}</strong>
            <small>AI GATEWAY</small>
          </span>
        </router-link>

        <div class="showcase-links">
          <a href="#why-us">{{ t('home.showcase.nav.whyUs') }}</a>
          <a href="#steps">{{ t('home.showcase.nav.steps') }}</a>
          <a href="#tools">{{ t('home.showcase.nav.apps') }}</a>
          <a href="#faq">{{ t('home.showcase.nav.faq') }}</a>
          <a href="#pricing">{{ t('home.showcase.nav.pricing') }}</a>
        </div>

        <div class="showcase-actions">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="showcase-icon-button"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <button
            type="button"
            class="showcase-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="emit('toggleTheme')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="showcase-login">
            <span v-if="isAuthenticated" class="login-avatar">{{ userInitial }}</span>
            <span>{{ isAuthenticated ? t('home.dashboard') : t('home.login') }}</span>
            <Icon name="arrowRight" size="xs" :stroke-width="2" />
          </router-link>
        </div>
      </nav>
    </header>

    <main>
      <section class="showcase-hero" aria-labelledby="showcase-title">
        <div class="hero-copy">
          <div class="hero-eyebrow"><span class="status-pulse" aria-hidden="true"></span>{{ t('home.showcase.eyebrow') }}</div>
          <h1 id="showcase-title">
            <span class="title-prefix">{{ t('home.showcase.titlePrefix') }}</span>
            <span>{{ t('home.showcase.titleHighlight') }}</span>
          </h1>
          <p class="hero-description">{{ t('home.showcase.description') }}</p>
          <div class="hero-cta">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="showcase-primary-button">
              <Icon name="bolt" size="sm" />
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <router-link to="/model-plaza" class="showcase-secondary-button">
              <Icon name="dollar" size="sm" />{{ t('home.showcase.nav.pricing') }}
            </router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="showcase-text-link">
              {{ t('home.showcase.readDocs') }} <Icon name="arrowRight" size="sm" />
            </a>
          </div>
          <div class="hero-proof" aria-label="Platform highlights">
            <span><Icon name="checkCircle" size="sm" />{{ t('home.showcase.proof.gateway') }}</span>
            <span><Icon name="checkCircle" size="sm" />{{ t('home.showcase.proof.failover') }}</span>
            <span><Icon name="checkCircle" size="sm" />{{ t('home.showcase.proof.billing') }}</span>
          </div>
        </div>

        <div class="hero-visual">
          <div class="terminal-container tool-visual" aria-label="Supported AI tools preview">
            <div class="tool-visual-head">
              <div class="window-dots" aria-hidden="true"><i></i><i></i><i></i></div>
              <span>{{ t('home.showcase.apps.previewTitle') }}</span>
              <span class="preview-live"><i></i>{{ t('home.showcase.apps.previewStatus') }}</span>
            </div>
            <div class="tool-visual-body">
              <aside class="tool-list" aria-label="Supported tools">
                <button
                  v-for="(app, index) in toolCatalog"
                  :key="app.key"
                  type="button"
                  class="tool-list-item"
                  :class="{ active: activeToolIndex === index }"
                  :aria-pressed="activeToolIndex === index"
                  @click="activeToolIndex = index"
                >
                  <span class="tool-list-icon"><OfficialToolIcon :tool="app.tool" :size="24" /></span>
                  <span class="tool-list-copy"><strong>{{ app.name }}</strong><small>{{ t(`home.showcase.apps.items.${app.key}.maker`) }}</small></span>
                  <Icon name="chevronRight" size="xs" />
                </button>
              </aside>
              <div class="tool-preview">
                <div class="preview-topline"><span>SUB2API / WORKFLOW</span><b>200 OK</b></div>
                <div class="preview-featured">
                  <span class="preview-icon"><OfficialToolIcon :tool="activeTool.tool" :size="42" /></span>
                  <div><small>{{ t('home.showcase.apps.previewSelected') }}</small><h3>{{ activeTool.name }}</h3></div>
                  <span class="preview-check"><Icon name="check" size="sm" /></span>
                </div>
                <p>{{ t(`home.showcase.apps.items.${activeTool.key}.description`) }}</p>
                <div class="preview-tags"><span>API READY</span><span>ONE KEY</span><span>LOW LATENCY</span></div>
                <div class="preview-bars" aria-hidden="true"><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i></div>
                <div class="preview-foot"><span><i></i>{{ t('home.showcase.apps.previewConnected') }}</span><span>{{ t('home.showcase.apps.previewUpdated') }}</span></div>
              </div>
            </div>
          </div>
          <p class="visual-caption"><span></span>{{ t('home.showcase.visualCaption') }}</p>
        </div>
      </section>

      <section class="status-strip" aria-label="Platform status">
        <div class="status-item"><span class="status-index">01</span><span><small>{{ t('home.showcase.status.routing') }}</small><strong>GPT + CLAUDE</strong></span></div>
        <div class="status-item"><span class="status-index">02</span><span><small>{{ t('home.showcase.status.reliability') }}</small><strong>{{ t('home.showcase.status.failover') }}</strong></span></div>
        <div class="status-item"><span class="status-index">03</span><span><small>{{ t('home.showcase.status.usage') }}</small><strong>{{ t('home.showcase.status.payg') }}</strong></span></div>
        <div class="status-item"><span class="status-index status-index-live"><i></i></span><span><small>API STATUS</small><strong>OPERATIONAL</strong></span></div>
      </section>

      <section id="why-us" class="landing-section section-soft" aria-labelledby="why-us-title">
        <div class="section-heading centered">
          <span class="section-kicker">{{ t('home.showcase.whyUs.eyebrow') }}</span>
          <h2 id="why-us-title">{{ t('home.showcase.whyUs.title') }}</h2>
          <p>{{ t('home.showcase.whyUs.description') }}</p>
        </div>
        <div class="value-grid">
          <article v-for="(item, index) in whyUsItems" :key="item.key" class="value-card">
            <span class="value-number">0{{ index + 1 }}</span>
            <span class="value-icon"><Icon :name="item.icon" size="lg" /></span>
            <h3>{{ t(`home.showcase.whyUs.items.${item.key}.title`) }}</h3>
            <p>{{ t(`home.showcase.whyUs.items.${item.key}.description`) }}</p>
          </article>
        </div>
      </section>

      <section id="steps" class="landing-section steps-section" aria-labelledby="steps-title">
        <div class="section-heading centered">
          <span class="section-kicker">{{ t('home.showcase.steps.eyebrow') }}</span>
          <h2 id="steps-title">{{ t('home.showcase.steps.title') }}</h2>
          <p>{{ t('home.showcase.steps.description') }}</p>
        </div>
        <div class="steps-grid">
          <article v-for="(item, index) in stepItems" :key="item.key" class="step-card">
            <div class="step-top"><span class="step-number">0{{ index + 1 }}</span><Icon :name="item.icon" size="lg" /></div>
            <h3>{{ t(`home.showcase.steps.items.${item.key}.title`) }}</h3>
            <p>{{ t(`home.showcase.steps.items.${item.key}.description`) }}</p>
            <span class="step-line" aria-hidden="true"></span>
          </article>
        </div>
      </section>

      <section id="tools" class="landing-section apps-section section-soft" aria-labelledby="apps-title">
        <div class="section-heading section-heading-row">
          <div><span class="section-kicker">{{ t('home.showcase.apps.eyebrow') }}</span><h2 id="apps-title">{{ t('home.showcase.apps.title') }}</h2><p>{{ t('home.showcase.apps.description') }}</p></div>
          <span class="section-counter">{{ toolCatalog.length }} <small>{{ t('home.showcase.apps.supported') }}</small></span>
        </div>
        <div class="apps-grid" role="list" :aria-label="t('home.showcase.apps.title')">
          <article v-for="app in toolCatalog" :key="app.key" class="app-card">
            <div class="app-card-top"><span class="app-icon"><OfficialToolIcon :tool="app.tool" :size="34" /></span><span class="app-badge">{{ t('home.showcase.apps.website') }}</span></div>
            <h3>{{ app.name }}</h3>
            <p>{{ t(`home.showcase.apps.items.${app.key}.description`) }}</p>
            <div class="app-card-foot"><span>{{ t(`home.showcase.apps.items.${app.key}.maker`) }}</span><a :href="app.url" target="_blank" rel="noopener noreferrer" class="app-download"><Icon name="download" size="sm" />{{ t('home.showcase.apps.download') }}<Icon name="externalLink" size="xs" /></a></div>
          </article>
        </div>
      </section>

      <section id="pricing" class="pricing-section" aria-labelledby="pricing-title">
        <div class="pricing-copy"><span class="section-kicker">{{ t('home.showcase.pricing.eyebrow') }}</span><h2 id="pricing-title">{{ t('home.showcase.pricing.title') }}</h2><p>{{ t('home.showcase.pricing.description') }}</p><router-link to="/model-plaza" class="section-link">{{ t('home.showcase.pricing.action') }} <Icon name="arrowRight" size="sm" /></router-link></div>
        <div class="pricing-features"><span><Icon name="dollar" size="md" /><strong>{{ t('home.showcase.pricing.features.payAsYouGo') }}</strong></span><span><Icon name="chart" size="md" /><strong>{{ t('home.showcase.pricing.features.transparent') }}</strong></span><span><Icon name="shield" size="md" /><strong>{{ t('home.showcase.pricing.features.quotaControl') }}</strong></span><small>{{ t('home.showcase.pricing.note') }}</small></div>
      </section>

      <section class="landing-section audience-section" aria-labelledby="audience-title">
        <div class="audience-panel">
          <div class="audience-copy"><span class="section-kicker">{{ t('home.showcase.developerEnterprise.eyebrow') }}</span><h2 id="audience-title">{{ t('home.showcase.developerEnterprise.title') }}</h2><p>{{ t('home.showcase.developerEnterprise.description') }}</p><router-link :to="isAuthenticated ? dashboardPath : '/login'" class="showcase-primary-button">{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }} <Icon name="arrowRight" size="sm" /></router-link></div>
          <div class="audience-points"><article v-for="item in audienceItems" :key="item.key"><span class="audience-icon"><Icon :name="item.icon" size="md" /></span><div><h3>{{ t(`home.showcase.developerEnterprise.${item.key}.title`) }}</h3><p>{{ t(`home.showcase.developerEnterprise.${item.key}.description`) }}</p></div></article></div>
        </div>
      </section>

      <section id="faq" class="landing-section faq-section section-soft" aria-labelledby="faq-title">
        <div class="section-heading centered"><span class="section-kicker">{{ t('home.showcase.faq.eyebrow') }}</span><h2 id="faq-title">{{ t('home.showcase.faq.title') }}</h2><p>{{ t('home.showcase.faq.description') }}</p></div>
        <div class="faq-list"><details v-for="(item, index) in faqItems" :key="item.key" class="faq-item" :open="index === 0"><summary><span>{{ t(`home.showcase.faq.items.${item.key}.question`) }}</span><Icon name="chevronDown" size="sm" /></summary><p>{{ t(`home.showcase.faq.items.${item.key}.answer`) }}</p></details></div>
      </section>

      <section class="showcase-cta" aria-labelledby="cta-title"><div><span class="section-kicker">08 / CONNECT</span><h2 id="cta-title">{{ t('home.cta.title') }}</h2><p>{{ t('home.cta.description') }}</p></div><router-link :to="isAuthenticated ? dashboardPath : '/login'" class="showcase-primary-button">{{ isAuthenticated ? t('home.goToDashboard') : t('home.cta.button') }}<Icon name="arrowRight" size="sm" /></router-link></section>
    </main>

    <footer class="showcase-footer"><div class="footer-brand"><span class="brand-mark"><img :src="siteLogo || '/logo.svg'" :alt="siteName" /></span><div><strong>{{ siteName }}</strong><small>{{ t('home.footerTagline') }}</small></div></div><div class="footer-links"><router-link to="/model-plaza">{{ t('home.showcase.nav.pricing') }}</router-link><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a><a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a></div><p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p></footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, toRefs } from 'vue'
import { useI18n } from 'vue-i18n'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import OfficialToolIcon from './OfficialToolIcon.vue'

defineOptions({ name: 'HomeShowcase' })

const props = defineProps<{
  siteName: string
  siteLogo: string
  siteSubtitle: string
  docUrl: string
  isDark: boolean
  isAuthenticated: boolean
  dashboardPath: string
  userInitial: string
  githubUrl: string
  currentYear: number
}>()
const { siteName, siteLogo, docUrl, isDark, isAuthenticated, dashboardPath, userInitial, githubUrl, currentYear } = toRefs(props)
const emit = defineEmits<{ toggleTheme: [] }>()
const { t } = useI18n()

const toolCatalog = [
  { key: 'claude', name: 'Claude', tool: 'claude', url: 'https://claude.ai/download' },
  { key: 'codex', name: 'Codex', tool: 'codex', url: 'https://github.com/openai/codex' },
  { key: 'openclaw', name: 'OpenClaw', tool: 'openclaw', url: 'https://github.com/openclaw/openclaw' },
  { key: 'hermes', name: 'Hermes', tool: 'hermes', url: 'https://github.com/NousResearch/hermes-agent' },
  { key: 'deepseekHarness', name: 'DeepSeek Harness', tool: 'deepseek-harness', url: 'https://github.com/deepseek-ai/deepseek-harness' },
  { key: 'ccSwitch', name: 'CC Switch', tool: 'ccswitch', url: 'https://github.com/farion1231/cc-switch' }
] as const
const whyUsItems = [
  { key: 'unified', icon: 'key' },
  { key: 'resilient', icon: 'swap' },
  { key: 'transparent', icon: 'chart' },
  { key: 'private', icon: 'shield' }
] as const
const stepItems = [
  { key: 'account', icon: 'userPlus' },
  { key: 'configure', icon: 'key' },
  { key: 'launch', icon: 'bolt' }
] as const
const audienceItems = [
  { key: 'developer', icon: 'terminal' },
  { key: 'enterprise', icon: 'users' }
] as const
const faqItems = [
  { key: 'what' },
  { key: 'pricing' },
  { key: 'security' },
  { key: 'support' }
] as const

const activeToolIndex = ref(0)
let toolTimer: ReturnType<typeof setInterval> | undefined
const activeTool = computed(() => toolCatalog[activeToolIndex.value % toolCatalog.length])

onMounted(() => {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
  toolTimer = setInterval(() => {
    activeToolIndex.value = (activeToolIndex.value + 1) % toolCatalog.length
  }, 4200)
})

onUnmounted(() => {
  if (toolTimer) clearInterval(toolTimer)
})
</script>

<style scoped>
.home-showcase {
  --showcase-bg: #fbfcfe;
  --showcase-bg-soft: #f4f7fb;
  --showcase-ink: #101828;
  --showcase-muted: #667085;
  --showcase-line: #e4e7ec;
  --showcase-panel: #fff;
  --showcase-accent: #3ea4ec;
  --showcase-accent-dark: #1779bf;
  --showcase-violet: #7657d9;
  --showcase-shadow: rgba(16, 24, 40, .07);
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  color: var(--showcase-ink);
  background: var(--showcase-bg);
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
}
.home-showcase.is-dark { --showcase-bg: #0b111b; --showcase-bg-soft: #111a28; --showcase-ink: #f0f4fa; --showcase-muted: #9aa9bc; --showcase-line: rgba(165, 183, 207, .18); --showcase-panel: #111a28; --showcase-accent: #6db9f5; --showcase-accent-dark: #a6d8ff; --showcase-violet: #a795fa; --showcase-shadow: rgba(0, 0, 0, .24); }
.showcase-grid { position: absolute; inset: 0; pointer-events: none; opacity: 0; }
.showcase-header, .showcase-hero, .status-strip, .landing-section, .showcase-cta, .showcase-footer { position: relative; z-index: 1; }
.showcase-header { position: sticky; top: 0; padding: 14px clamp(18px, 4vw, 64px); border-bottom: 1px solid color-mix(in srgb, var(--showcase-line) 78%, transparent); background: color-mix(in srgb, var(--showcase-bg) 90%, transparent); backdrop-filter: blur(18px); }
.showcase-nav { display: flex; align-items: center; gap: 30px; max-width: 1180px; margin: 0 auto; }
.showcase-brand, .showcase-actions, .showcase-links, .hero-cta, .hero-proof, .status-item, .tool-visual-head, .preview-topline, .preview-foot, .app-card-top, .app-card-foot, .model-top, .model-foot, .section-heading-row, .showcase-footer, .footer-brand, .footer-links { display: flex; align-items: center; }
.showcase-brand { min-width: 0; gap: 11px; color: inherit; text-decoration: none; }
.brand-mark { display: grid; width: 40px; height: 40px; flex: 0 0 auto; place-items: center; border: 1px solid var(--showcase-line); border-radius: 10px; background: var(--showcase-panel); box-shadow: 0 8px 20px var(--showcase-shadow); }
.brand-mark img { width: 28px; height: 28px; object-fit: contain; }
.brand-copy { display: grid; min-width: 0; gap: 3px; }
.brand-copy strong { overflow: hidden; font-size: 14px; font-weight: 750; text-overflow: ellipsis; white-space: nowrap; }
.brand-copy small, .section-kicker, .app-badge, .section-counter, .step-number { color: var(--showcase-muted); font-size: 10px; font-weight: 700; letter-spacing: .08em; }
.brand-copy small { font-size: 8px; }
.showcase-links { gap: 23px; margin-left: auto; }
.showcase-links a, .showcase-footer a { color: var(--showcase-muted); font-size: 13px; text-decoration: none; transition: color .2s ease; }
.showcase-links a:hover, .showcase-footer a:hover { color: var(--showcase-accent-dark); }
.showcase-actions { gap: 7px; }
.showcase-icon-button { display: grid; width: 36px; height: 36px; place-items: center; border: 1px solid transparent; border-radius: 9px; color: var(--showcase-muted); background: transparent; cursor: pointer; transition: .2s ease; }
.showcase-icon-button:hover, .showcase-icon-button:focus-visible { border-color: var(--showcase-line); color: var(--showcase-accent-dark); background: var(--showcase-panel); outline: none; }
.showcase-login { display: inline-flex; align-items: center; gap: 7px; min-height: 38px; padding: 0 13px; border-radius: 9px; color: #fff; background: #101828; font-size: 12px; font-weight: 700; text-decoration: none; transition: .2s ease; }
.showcase-login:hover, .showcase-login:focus-visible { background: #26354c; outline: none; }
.login-avatar { display: grid; width: 21px; height: 21px; place-items: center; border-radius: 50%; color: #0a263b; background: #b9e1fb; font-size: 10px; }

.showcase-hero { max-width: 1180px; margin: 0 auto; padding: 82px clamp(18px, 4vw, 64px) 76px; text-align: center; }
.hero-copy { max-width: 820px; margin: 0 auto; }
.hero-eyebrow { display: inline-flex; align-items: center; gap: 8px; margin-bottom: 22px; color: var(--showcase-accent-dark); font-size: 11px; font-weight: 750; letter-spacing: .08em; }
.status-pulse, .preview-live i, .status-index-live i, .preview-foot i { display: inline-block; width: 7px; height: 7px; border-radius: 50%; background: #28a66f; box-shadow: 0 0 0 4px rgba(40, 166, 111, .12); }
.hero-copy h1 { max-width: none; margin: 0 auto; font-size: clamp(38px, 5.2vw, 64px); font-weight: 760; line-height: 1.08; letter-spacing: -.025em; white-space: nowrap; }
.hero-copy h1 span { display: inline; color: var(--showcase-accent); }
.hero-copy h1 .title-prefix { margin-right: .22em; color: var(--showcase-ink); }
.hero-description { max-width: 650px; margin: 13px auto 0; color: var(--showcase-muted); font-size: 15px; line-height: 1.8; }
.hero-cta { justify-content: center; flex-wrap: wrap; gap: 10px; margin-top: 29px; }
.showcase-primary-button, .showcase-secondary-button { display: inline-flex; align-items: center; justify-content: center; gap: 8px; min-height: 46px; padding: 0 18px; border-radius: 9px; font-size: 13px; font-weight: 750; text-decoration: none; transition: transform .2s ease, box-shadow .2s ease, background .2s ease; }
.showcase-primary-button { color: #fff; background: var(--showcase-accent); box-shadow: 0 10px 24px rgba(62, 164, 236, .22); }
.showcase-primary-button:hover { transform: translateY(-2px); background: var(--showcase-accent-dark); box-shadow: 0 14px 28px rgba(62, 164, 236, .28); }
.showcase-secondary-button { border: 1px solid var(--showcase-line); color: var(--showcase-ink); background: var(--showcase-panel); }
.showcase-secondary-button:hover { border-color: var(--showcase-accent); color: var(--showcase-accent-dark); transform: translateY(-2px); }
.showcase-text-link, .section-link, .app-link { display: inline-flex; align-items: center; gap: 6px; color: var(--showcase-accent-dark); font-size: 13px; font-weight: 750; text-decoration: none; }
.showcase-text-link:hover, .section-link:hover, .app-link:hover { color: var(--showcase-ink); }
.hero-proof { justify-content: center; flex-wrap: wrap; gap: 12px 22px; margin-top: 25px; color: var(--showcase-muted); font-size: 12px; }
.hero-proof span { display: inline-flex; align-items: center; gap: 6px; }
.hero-proof svg { color: #28a66f; }

.hero-visual { max-width: 980px; margin: 58px auto 0; text-align: left; }
.tool-visual { overflow: hidden; border: 1px solid #d9e2ed; border-radius: 15px; background: #fff; box-shadow: 0 24px 65px rgba(45, 71, 105, .15); }
.is-dark .tool-visual { border-color: var(--showcase-line); background: #101927; box-shadow: 0 24px 65px rgba(0, 0, 0, .3); }
.tool-visual-head { justify-content: space-between; min-height: 51px; padding: 0 19px; border-bottom: 1px solid var(--showcase-line); color: var(--showcase-muted); font-size: 10px; font-weight: 750; letter-spacing: .08em; }
.window-dots { display: flex; gap: 5px; }
.window-dots i { width: 7px; height: 7px; border-radius: 50%; background: #cbd5e1; }
.window-dots i:nth-child(1) { background: #ff8c83; }.window-dots i:nth-child(2) { background: #f8c75d; }.window-dots i:nth-child(3) { background: #66c987; }
.preview-live { display: inline-flex; align-items: center; gap: 7px; color: #28a66f; }
.tool-visual-body { display: grid; grid-template-columns: minmax(235px, .82fr) minmax(0, 1.18fr); min-height: 328px; }
.tool-list { padding: 18px 13px; border-right: 1px solid var(--showcase-line); background: #f7faff; }
.is-dark .tool-list { background: #0d1623; }
.tool-list-item { display: flex; align-items: center; width: 100%; min-height: 47px; gap: 9px; padding: 7px 9px; border: 1px solid transparent; border-radius: 9px; color: var(--showcase-muted); background: transparent; text-align: left; cursor: pointer; transition: .2s ease; }
.tool-list-item + .tool-list-item { margin-top: 4px; }
.tool-list-item:hover, .tool-list-item.active { border-color: color-mix(in srgb, var(--showcase-accent) 35%, var(--showcase-line)); color: var(--showcase-ink); background: var(--showcase-panel); box-shadow: 0 7px 17px var(--showcase-shadow); }
.tool-list-icon, .app-icon { display: grid; flex: 0 0 auto; place-items: center; width: 37px; height: 37px; border: 1px solid var(--showcase-line); border-radius: 9px; background: var(--showcase-panel); }
.tool-list-copy { display: grid; min-width: 0; gap: 3px; }.tool-list-copy strong { overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }.tool-list-copy small { overflow: hidden; color: var(--showcase-muted); font-size: 9px; text-overflow: ellipsis; white-space: nowrap; }.tool-list-item > svg { margin-left: auto; color: var(--showcase-muted); }
.tool-preview { display: flex; flex-direction: column; min-width: 0; padding: 25px 29px 20px; }
.preview-topline, .preview-foot { justify-content: space-between; color: var(--showcase-muted); font-size: 9px; font-weight: 750; letter-spacing: .08em; }.preview-topline b { color: #28a66f; font-weight: 750; }
.preview-featured { display: flex; align-items: center; gap: 12px; margin-top: 31px; }.preview-icon { display: grid; width: 58px; height: 58px; place-items: center; border: 1px solid color-mix(in srgb, var(--showcase-accent) 30%, var(--showcase-line)); border-radius: 13px; background: color-mix(in srgb, var(--showcase-accent) 9%, var(--showcase-panel)); }.preview-featured small { color: var(--showcase-muted); font-size: 9px; font-weight: 700; letter-spacing: .08em; }.preview-featured h3 { margin: 4px 0 0; font-size: 24px; letter-spacing: -.02em; }.preview-check { display: grid; width: 27px; height: 27px; place-items: center; margin-left: auto; border-radius: 50%; color: #fff; background: #28a66f; }.tool-preview > p { max-width: 420px; margin: 19px 0 0; color: var(--showcase-muted); font-size: 13px; line-height: 1.7; }.preview-tags { display: flex; flex-wrap: wrap; gap: 7px; margin-top: 17px; }.preview-tags span { padding: 5px 8px; border: 1px solid var(--showcase-line); border-radius: 5px; color: var(--showcase-muted); font-size: 8px; font-weight: 750; letter-spacing: .06em; }.preview-bars { display: flex; align-items: end; gap: 5px; height: 34px; margin-top: auto; }.preview-bars i { width: 10px; height: 35%; border-radius: 3px 3px 0 0; background: color-mix(in srgb, var(--showcase-accent) 45%, var(--showcase-line)); animation: preview-bars 2.5s ease-in-out infinite alternate; }.preview-bars i:nth-child(2) { height: 58%; animation-delay: -.3s; }.preview-bars i:nth-child(3) { height: 42%; animation-delay: -.6s; }.preview-bars i:nth-child(4) { height: 78%; animation-delay: -.9s; }.preview-bars i:nth-child(5) { height: 53%; animation-delay: -1.2s; }.preview-bars i:nth-child(6) { height: 88%; animation-delay: -1.5s; }.preview-bars i:nth-child(7) { height: 66%; animation-delay: -1.8s; }.preview-bars i:nth-child(8) { height: 96%; animation-delay: -2.1s; }.preview-foot { margin-top: 17px; padding-top: 12px; border-top: 1px solid var(--showcase-line); font-size: 9px; letter-spacing: .03em; }.preview-foot span:first-child { display: inline-flex; align-items: center; gap: 7px; }.visual-caption { display: flex; align-items: center; gap: 7px; margin: 14px 0 0 3px; color: var(--showcase-muted); font-size: 11px; }.visual-caption > span { width: 6px; height: 6px; border-radius: 50%; background: var(--showcase-violet); }

.status-strip { display: grid; grid-template-columns: repeat(4, 1fr); max-width: 1060px; margin: 0 auto; border: 1px solid var(--showcase-line); border-radius: 12px; background: var(--showcase-panel); box-shadow: 0 12px 30px var(--showcase-shadow); }.status-item { gap: 12px; min-height: 78px; padding: 13px 19px; border-right: 1px solid var(--showcase-line); }.status-item:last-child { border-right: 0; }.status-item > span:last-child { display: grid; gap: 5px; }.status-item small { color: var(--showcase-muted); font-size: 9px; font-weight: 700; letter-spacing: .07em; }.status-item strong { font-size: 12px; }.status-index { color: var(--showcase-accent-dark); font-size: 10px; font-weight: 750; }.status-index-live i { display: block; }

.landing-section { max-width: 1060px; margin: 0 auto; padding: 104px 0; }.section-soft { max-width: none; padding-right: max(18px, calc((100vw - 1060px) / 2)); padding-left: max(18px, calc((100vw - 1060px) / 2)); background: var(--showcase-bg-soft); }.section-heading { max-width: 650px; margin-bottom: 39px; }.section-heading.centered { margin-right: auto; margin-left: auto; text-align: center; }.section-kicker { display: block; margin-bottom: 13px; color: var(--showcase-accent-dark); }.section-heading h2, .audience-copy h2, .showcase-cta h2 { margin: 0; font-size: clamp(30px, 4vw, 47px); font-weight: 740; line-height: 1.08; letter-spacing: -.035em; text-wrap: balance; }.section-heading p, .audience-copy > p, .showcase-cta p { margin: 14px 0 0; color: var(--showcase-muted); font-size: 14px; line-height: 1.8; }
.value-grid, .steps-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 13px; }.value-card, .step-card, .app-card { border: 1px solid var(--showcase-line); border-radius: 12px; background: var(--showcase-panel); box-shadow: 0 11px 28px var(--showcase-shadow); }.value-card { position: relative; min-height: 224px; padding: 24px 22px; }.value-number { position: absolute; top: 24px; right: 22px; color: var(--showcase-muted); font-size: 10px; font-weight: 700; }.value-icon { display: grid; width: 45px; height: 45px; place-items: center; border-radius: 11px; color: var(--showcase-accent-dark); background: #eaf6ff; }.is-dark .value-icon { background: rgba(109, 185, 245, .12); }.value-card h3, .step-card h3, .app-card h3 { margin: 22px 0 8px; font-size: 17px; }.value-card p, .step-card p, .app-card p { margin: 0; color: var(--showcase-muted); font-size: 12px; line-height: 1.7; }
.steps-section { padding-top: 102px; }.steps-grid { grid-template-columns: repeat(3, 1fr); }.step-card { position: relative; min-height: 223px; padding: 24px 23px; overflow: hidden; }.step-top { display: flex; align-items: center; justify-content: space-between; color: var(--showcase-accent-dark); }.step-number { color: var(--showcase-accent-dark); font-size: 12px; }.step-card h3 { margin-top: 33px; }.step-line { position: absolute; right: -15%; bottom: 0; left: 23px; height: 3px; background: var(--showcase-accent); transform: scaleX(.6); transform-origin: left; transition: transform .3s ease; }.step-card:hover .step-line { transform: scaleX(1); }
.section-heading-row { justify-content: space-between; align-items: end; gap: 25px; }.section-heading-row .section-heading, .section-heading-row > div { margin-bottom: 0; }.section-counter { display: grid; flex: 0 0 auto; gap: 4px; color: var(--showcase-accent-dark); font-size: 28px; letter-spacing: -.04em; }.section-counter small { color: var(--showcase-muted); font-size: 10px; letter-spacing: .05em; }.apps-section { padding-top: 104px; }.apps-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 13px; }.app-card { display: flex; min-height: 250px; flex-direction: column; padding: 19px; transition: transform .2s ease, border-color .2s ease; }.app-card:hover, .value-card:hover, .step-card:hover { border-color: color-mix(in srgb, var(--showcase-accent) 48%, var(--showcase-line)); transform: translateY(-3px); }.app-card-top { justify-content: space-between; }.app-icon { width: 48px; height: 48px; border-radius: 12px; }.app-badge { padding: 5px 7px; border-radius: 5px; color: #218454; background: #e9f8f0; font-size: 8px; letter-spacing: .04em; }.app-card h3 { margin-top: 20px; }.app-card-foot { align-items: stretch; flex-direction: column; gap: 11px; margin-top: auto; padding-top: 16px; border-top: 1px solid var(--showcase-line); }.app-card-foot > span { overflow: hidden; color: var(--showcase-muted); font-size: 9px; text-overflow: ellipsis; white-space: nowrap; }.app-download { display: inline-flex; align-items: center; justify-content: center; gap: 7px; min-height: 39px; padding: 0 12px; border-radius: 8px; color: #fff; background: var(--showcase-accent); font-size: 11px; font-weight: 750; text-decoration: none; transition: background .2s ease, transform .2s ease; }.app-download:hover, .app-download:focus-visible { background: var(--showcase-accent-dark); transform: translateY(-1px); outline: none; }
.pricing-section { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr); align-items: center; gap: 65px; max-width: 1060px; margin: 0 auto; padding: 22px 0 105px; }.pricing-copy { max-width: 500px; }.pricing-copy h2 { margin: 0; font-size: clamp(30px, 4vw, 45px); font-weight: 740; line-height: 1.08; letter-spacing: -.035em; }.pricing-copy p { margin: 14px 0 0; color: var(--showcase-muted); font-size: 14px; line-height: 1.8; }.pricing-copy .section-link { margin-top: 22px; }.pricing-features { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }.pricing-features span { display: grid; gap: 18px; min-height: 116px; padding: 17px; border: 1px solid var(--showcase-line); border-radius: 11px; background: var(--showcase-panel); box-shadow: 0 11px 25px var(--showcase-shadow); }.pricing-features span svg { color: var(--showcase-accent-dark); }.pricing-features strong { font-size: 12px; }.pricing-features small { grid-column: 1 / -1; margin-top: 5px; color: var(--showcase-muted); font-size: 11px; }
.audience-section { padding-top: 16px; }.audience-panel { display: grid; grid-template-columns: minmax(0, .9fr) minmax(0, 1.1fr); gap: 58px; padding: 56px; border-radius: 16px; color: #fff; background: #15243a; box-shadow: 0 22px 48px rgba(16, 24, 40, .18); }.audience-panel .section-kicker { color: #8cccf6; }.audience-copy h2 { color: #fff; }.audience-copy > p { color: #b9c7d8; }.audience-copy .showcase-primary-button { margin-top: 27px; color: #0c2940; background: #b9e1fb; box-shadow: none; }.audience-copy .showcase-primary-button:hover { background: #d7efff; }.audience-points { display: grid; gap: 11px; align-content: center; }.audience-points article { display: flex; gap: 13px; padding: 19px; border: 1px solid rgba(217, 231, 247, .18); border-radius: 11px; background: rgba(255, 255, 255, .06); }.audience-icon { display: grid; width: 39px; height: 39px; flex: 0 0 auto; place-items: center; border-radius: 9px; color: #b9e1fb; background: rgba(185, 225, 251, .12); }.audience-points h3 { margin: 0; font-size: 15px; }.audience-points p { margin: 5px 0 0; color: #b9c7d8; font-size: 12px; line-height: 1.65; }
.faq-section { padding-top: 105px; }.faq-list { max-width: 780px; margin: 0 auto; }.faq-item { border-top: 1px solid var(--showcase-line); }.faq-item:last-child { border-bottom: 1px solid var(--showcase-line); }.faq-item summary { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 20px 4px; color: var(--showcase-ink); font-size: 15px; font-weight: 700; list-style: none; cursor: pointer; }.faq-item summary::-webkit-details-marker { display: none; }.faq-item summary svg { flex: 0 0 auto; color: var(--showcase-muted); transition: transform .2s ease; }.faq-item[open] summary svg { transform: rotate(180deg); }.faq-item p { max-width: 680px; margin: -4px 30px 19px 4px; color: var(--showcase-muted); font-size: 13px; line-height: 1.8; }
.showcase-cta { display: flex; align-items: center; justify-content: space-between; gap: 28px; max-width: 1060px; margin: 0 auto 76px; padding: 38px 43px; border: 1px solid var(--showcase-line); border-radius: 14px; background: var(--showcase-panel); box-shadow: 0 15px 34px var(--showcase-shadow); }.showcase-cta h2 { font-size: clamp(25px, 3vw, 37px); }.showcase-cta p { max-width: 570px; }
.showcase-footer { flex-wrap: wrap; gap: 18px 35px; max-width: 1060px; margin: 0 auto; padding: 24px 0 34px; border-top: 1px solid var(--showcase-line); color: var(--showcase-muted); font-size: 11px; }.footer-brand { gap: 9px; }.footer-brand .brand-mark { width: 31px; height: 31px; border-radius: 8px; }.footer-brand .brand-mark img { width: 21px; height: 21px; }.footer-brand div:last-child { display: grid; gap: 3px; }.footer-brand strong { color: var(--showcase-ink); font-size: 12px; }.footer-brand small { font-size: 10px; }.footer-links { gap: 20px; margin-left: auto; }.showcase-footer > p { width: 100%; margin: 0; }

@keyframes preview-bars { from { opacity: .48; transform: scaleY(.75); } to { opacity: 1; transform: scaleY(1); } }
@media (max-width: 980px) { .showcase-links { display: none; }.showcase-hero { padding-top: 66px; }.tool-visual-body { grid-template-columns: minmax(210px, .9fr) minmax(0, 1.1fr); }.landing-section { margin-right: 18px; margin-left: 18px; }.section-soft { margin-right: 0; margin-left: 0; }.value-grid { grid-template-columns: repeat(2, 1fr); }.apps-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.pricing-section { gap: 35px; margin-right: 18px; margin-left: 18px; }.audience-panel { padding: 42px; }.showcase-footer { margin-right: 18px; margin-left: 18px; } }
@media (max-width: 680px) { .showcase-header { padding: 11px 14px; }.showcase-nav { gap: 10px; }.brand-mark { width: 36px; height: 36px; }.brand-mark img { width: 25px; height: 25px; }.brand-copy { max-width: min(37vw, 170px); }.brand-copy small, .showcase-actions > :deep(.locale-switcher) { display: none; }.showcase-actions { gap: 3px; }.showcase-icon-button { width: 33px; height: 33px; }.showcase-login { min-height: 33px; padding: 0 9px; font-size: 11px; }.showcase-login svg { display: none; }.showcase-hero { padding: 54px 18px 57px; }.hero-copy h1 { font-size: clamp(38px, 12vw, 54px); white-space: normal; overflow-wrap: anywhere; }.hero-description { overflow-wrap: anywhere; }.hero-cta { align-items: stretch; flex-direction: column; }.showcase-primary-button, .showcase-secondary-button, .showcase-text-link { width: 100%; }.hero-proof { align-items: flex-start; flex-direction: column; gap: 9px; width: fit-content; margin-right: auto; margin-left: auto; text-align: left; }.hero-visual { margin-top: 39px; }.tool-visual-body { grid-template-columns: 1fr; }.tool-list { display: grid; grid-template-columns: repeat(2, 1fr); gap: 5px; padding: 11px; border-right: 0; border-bottom: 1px solid var(--showcase-line); }.tool-list-item { min-height: 43px; padding: 5px 6px; }.tool-list-item + .tool-list-item { margin-top: 0; }.tool-list-item > svg { display: none; }.tool-list-icon { width: 30px; height: 30px; }.tool-list-copy strong { font-size: 10px; }.tool-list-copy small { font-size: 8px; }.tool-preview { min-height: 255px; padding: 21px 18px 18px; }.preview-featured { margin-top: 23px; }.preview-featured h3 { font-size: 21px; }.status-strip { grid-template-columns: repeat(2, 1fr); margin: 0 18px; }.status-item { min-height: 66px; padding: 11px 10px; border-bottom: 1px solid var(--showcase-line); }.status-item:nth-child(2n) { border-right: 0; }.status-item:nth-child(3), .status-item:nth-child(4) { border-bottom: 0; }.landing-section { padding: 75px 0; }.section-soft { padding-right: 18px; padding-left: 18px; }.section-heading-row { align-items: flex-start; }.section-counter { font-size: 23px; }.value-grid, .steps-grid, .apps-grid { grid-template-columns: 1fr; }.apps-grid { gap: 11px; }.app-card { min-height: 232px; }.steps-section, .faq-section { padding-top: 75px; }.pricing-section { grid-template-columns: 1fr; gap: 30px; margin-right: 18px; margin-left: 18px; padding-top: 0; padding-bottom: 75px; }.pricing-features { grid-template-columns: repeat(3, 1fr); }.pricing-features span { min-height: 100px; padding: 13px; }.pricing-features strong { font-size: 10px; }.pricing-features small { font-size: 10px; }.audience-section { padding-top: 0; }.audience-panel { grid-template-columns: 1fr; gap: 31px; padding: 28px 21px; border-radius: 13px; }.audience-copy .showcase-primary-button { width: 100%; }.showcase-cta { align-items: stretch; flex-direction: column; margin: 0 18px 61px; padding: 28px 21px; }.showcase-footer { align-items: flex-start; flex-direction: column; margin-right: 18px; margin-left: 18px; padding-bottom: 25px; }.footer-links { margin-left: 0; }.showcase-footer > p { width: auto; } }
@media (prefers-reduced-motion: reduce) { .preview-bars i { animation: none; }.showcase-primary-button, .showcase-secondary-button, .tool-list-item, .app-card, .value-card, .step-card { transition: none; } }
</style>

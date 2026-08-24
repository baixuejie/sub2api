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
          <a href="#capabilities">{{ t('home.showcase.capabilities') }}</a>
          <a href="#tools">{{ t('home.showcase.toolTitle') }}</a>
          <a href="#models">{{ t('home.showcase.engines') }}</a>
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
            {{ t('home.showcase.titlePrefix') }}
            <span>{{ siteName }}</span>
          </h1>
          <p class="hero-subtitle">{{ siteSubtitle }}</p>
          <p class="hero-description">{{ t('home.showcase.description') }}</p>
          <div class="hero-cta">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="showcase-primary-button">
              <Icon name="bolt" size="sm" />
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="showcase-secondary-button">
              <Icon name="book" size="sm" />{{ t('home.showcase.readDocs') }}
            </a>
          </div>
          <div class="hero-metrics" aria-label="Platform capabilities">
            <span><strong>01</strong>{{ t('home.showcase.proof.gateway') }}</span>
            <span><strong>02</strong>{{ t('home.showcase.proof.failover') }}</span>
            <span><strong>03</strong>{{ t('home.showcase.proof.billing') }}</span>
          </div>
        </div>

        <div class="hero-visual">
          <div class="terminal-container route-visual" aria-label="AI routing preview">
            <div class="route-visual-top">
              <span class="route-label">LIVE ROUTING</span>
              <span class="route-online"><i></i>{{ t('home.showcase.status.gateway') }} · READY</span>
            </div>
            <div class="route-graph">
              <div class="route-line route-line-main" aria-hidden="true"><span></span></div>
              <div class="route-line route-line-gpt" aria-hidden="true"><span></span></div>
              <div class="route-line route-line-claude" aria-hidden="true"><span></span></div>

              <div class="route-node route-node-ingress">
                <span class="route-node-icon"><Icon name="terminal" size="sm" /></span>
                <span><small>REQUEST</small><strong>/v1/chat/completions</strong></span>
              </div>
              <div class="route-node route-node-core">
                <span class="core-orbit" aria-hidden="true"></span>
                <span class="route-node-icon core-icon"><Icon name="cpu" size="sm" /></span>
                <span><small>SUB2API</small><strong>SMART ROUTER</strong></span>
              </div>
              <button
                type="button"
                class="route-node route-node-engine route-node-gpt"
                :class="{ active: activeEngine === 'gpt' }"
                :aria-pressed="activeEngine === 'gpt'"
                @click="activeEngine = 'gpt'"
              >
                <span class="route-brand"><OfficialToolIcon tool="codex" :size="27" /></span>
                <span><small>ENGINE A</small><strong>GPT</strong></span>
                <Icon name="chevronRight" size="xs" />
              </button>
              <button
                type="button"
                class="route-node route-node-engine route-node-claude"
                :class="{ active: activeEngine === 'claude' }"
                :aria-pressed="activeEngine === 'claude'"
                @click="activeEngine = 'claude'"
              >
                <span class="route-brand"><OfficialToolIcon tool="claude" :size="27" /></span>
                <span><small>ENGINE B</small><strong>CLAUDE</strong></span>
                <Icon name="chevronRight" size="xs" />
              </button>
            </div>
            <div class="route-visual-bottom">
              <span><i class="route-signal"></i>{{ activeEngine === 'gpt' ? 'GPT' : 'CLAUDE' }} · {{ t('home.showcase.responseReady') }}</span>
              <span><b>200</b> OK</span>
              <span>SECURE <Icon name="lock" size="xs" /></span>
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

      <section id="capabilities" class="capability-section" aria-labelledby="capability-title">
        <div class="section-heading">
          <span class="section-kicker">01 / PLATFORM</span>
          <h2 id="capability-title">{{ t('home.showcase.capabilityTitle') }}</h2>
          <p>{{ t('home.showcase.capabilityDescription') }}</p>
        </div>
        <div class="capability-rail">
          <article class="capability-item"><span class="capability-index">A</span><Icon name="server" size="lg" /><h3>{{ t('home.features.unifiedGateway') }}</h3><p>{{ t('home.features.unifiedGatewayDesc') }}</p></article>
          <article class="capability-item"><span class="capability-index">B</span><Icon name="swap" size="lg" /><h3>{{ t('home.features.multiAccount') }}</h3><p>{{ t('home.features.multiAccountDesc') }}</p></article>
          <article class="capability-item"><span class="capability-index">C</span><Icon name="chart" size="lg" /><h3>{{ t('home.features.balanceQuota') }}</h3><p>{{ t('home.features.balanceQuotaDesc') }}</p></article>
        </div>
      </section>

      <section id="tools" class="tool-section" aria-labelledby="tool-title">
        <div class="tool-copy">
          <span class="section-kicker">02 / TOOL ECOSYSTEM</span>
          <h2 id="tool-title">{{ t('home.showcase.toolTitle') }}</h2>
          <p>{{ t('home.showcase.toolDescription') }}</p>
        </div>
        <div class="tool-rail" role="list" aria-label="Claude, Codex, OpenClaw, Hermes, DeepSeek Harness and CC Switch">
          <div class="tool-rail-line" aria-hidden="true"><span></span></div>
          <div class="tool-item" role="listitem"><span class="tool-icon-frame"><OfficialToolIcon tool="claude" :size="32" /></span><span><strong>Claude</strong><small>ANTHROPIC</small></span><i class="tool-status"></i></div>
          <div class="tool-item" role="listitem"><span class="tool-icon-frame"><OfficialToolIcon tool="codex" :size="32" /></span><span><strong>Codex</strong><small>OPENAI</small></span><i class="tool-status"></i></div>
          <div class="tool-item" role="listitem"><span class="tool-icon-frame"><OfficialToolIcon tool="openclaw" :size="32" /></span><span><strong>OpenClaw</strong><small>OPEN SOURCE</small></span><i class="tool-status"></i></div>
          <div class="tool-item" role="listitem"><span class="tool-icon-frame"><OfficialToolIcon tool="hermes" :size="32" /></span><span><strong>Hermes</strong><small>NOUS RESEARCH</small></span><i class="tool-status"></i></div>
          <div class="tool-item" role="listitem"><span class="tool-icon-frame"><OfficialToolIcon tool="deepseek-harness" :size="32" /></span><span><strong>DeepSeek Harness</strong><small>DEEPSEEK AI</small></span><i class="tool-status"></i></div>
          <div class="tool-item" role="listitem"><span class="tool-icon-frame"><OfficialToolIcon tool="ccswitch" :size="32" /></span><span><strong>CC Switch</strong><small>COMMUNITY</small></span><i class="tool-status"></i></div>
        </div>
      </section>

      <section id="models" class="engine-section" aria-labelledby="engine-title">
        <div class="engine-intro">
          <span class="section-kicker">03 / MODEL ACCESS</span>
          <h2 id="engine-title">{{ t('home.showcase.engineTitle') }}</h2>
          <p>{{ t('home.showcase.engineDescription') }}</p>
          <router-link to="/model-plaza" class="inline-link">{{ t('home.showcase.exploreModels') }} <Icon name="arrowRight" size="sm" /></router-link>
        </div>
        <div class="engine-cards">
          <article class="engine-card engine-card-gpt"><div class="engine-card-head"><span class="engine-large-logo"><OfficialToolIcon tool="codex" :size="30" /></span><span class="engine-label">ENGINE A</span><span class="engine-live"><i></i>ACTIVE</span></div><h3>GPT</h3><p>{{ t('home.showcase.gptDescription') }}</p><div class="engine-card-foot"><span>GENERAL PURPOSE</span><Icon name="arrowUp" size="sm" /></div></article>
          <article class="engine-card engine-card-claude"><div class="engine-card-head"><span class="engine-large-logo"><OfficialToolIcon tool="claude" :size="30" /></span><span class="engine-label">ENGINE B</span><span class="engine-live"><i></i>ACTIVE</span></div><h3>Claude</h3><p>{{ t('home.showcase.claudeDescription') }}</p><div class="engine-card-foot"><span>REASONING &amp; WRITING</span><Icon name="arrowUp" size="sm" /></div></article>
        </div>
      </section>

      <section class="showcase-cta" aria-labelledby="cta-title">
        <div><span class="section-kicker">04 / CONNECT</span><h2 id="cta-title">{{ t('home.cta.title') }}</h2><p>{{ t('home.cta.description') }}</p></div>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="showcase-primary-button">{{ isAuthenticated ? t('home.goToDashboard') : t('home.cta.button') }}<Icon name="arrowRight" size="sm" /></router-link>
      </section>
    </main>

    <footer class="showcase-footer"><p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p><div><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a><a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a></div></footer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, toRefs } from 'vue'
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
const { siteName, siteLogo, siteSubtitle, docUrl, isDark, isAuthenticated, dashboardPath, userInitial, githubUrl, currentYear } = toRefs(props)
const emit = defineEmits<{ toggleTheme: [] }>()
const { t } = useI18n()
const activeEngine = ref<'gpt' | 'claude'>('gpt')
let routeTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    routeTimer = setInterval(() => { activeEngine.value = activeEngine.value === 'gpt' ? 'claude' : 'gpt' }, 5000)
  }
})

onUnmounted(() => { if (routeTimer) clearInterval(routeTimer) })
</script>

<style scoped>
.home-showcase {
  --showcase-bg: #f4f7fa;
  --showcase-bg-soft: #e9eef4;
  --showcase-ink: #101827;
  --showcase-muted: #637184;
  --showcase-line: #cdd7e2;
  --showcase-panel: rgba(255, 255, 255, .78);
  --showcase-accent: #0b9f9a;
  --showcase-violet: #6656d9;
  --showcase-shadow: rgba(18, 36, 58, .11);
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  color: var(--showcase-ink);
  background: var(--showcase-bg);
}
.home-showcase.is-dark { --showcase-bg: #06090f; --showcase-bg-soft: #0c121b; --showcase-ink: #e9f1fb; --showcase-muted: #8c9bad; --showcase-line: rgba(152, 177, 205, .18); --showcase-panel: rgba(13, 20, 31, .9); --showcase-accent: #45e6d1; --showcase-violet: #8b7cff; --showcase-shadow: rgba(0, 0, 0, .32); }
.showcase-grid { position: absolute; inset: 0; pointer-events: none; opacity: .32; background-image: linear-gradient(to right, var(--showcase-line) 1px, transparent 1px), linear-gradient(to bottom, var(--showcase-line) 1px, transparent 1px); background-size: 64px 64px; mask-image: linear-gradient(to bottom, black, transparent 55%); }
.showcase-header, .showcase-hero, .status-strip, .capability-section, .tool-section, .engine-section, .showcase-cta, .showcase-footer { position: relative; z-index: 1; }
.showcase-header { position: sticky; top: 0; padding: 15px clamp(18px, 4vw, 64px); border-bottom: 1px solid var(--showcase-line); background: color-mix(in srgb, var(--showcase-bg) 88%, transparent); backdrop-filter: blur(18px); }
.showcase-nav { display: flex; align-items: center; gap: 28px; max-width: 1240px; margin: 0 auto; }
.showcase-brand, .showcase-actions, .showcase-links, .hero-cta, .hero-metrics, .route-online, .route-visual-bottom, .status-item, .engine-card-head, .engine-card-foot, .showcase-footer > div { display: flex; align-items: center; }
.showcase-brand { min-width: 0; gap: 11px; color: inherit; text-decoration: none; }
.brand-mark { display: grid; width: 40px; height: 40px; flex: 0 0 auto; place-items: center; border: 1px solid var(--showcase-line); border-radius: 4px; background: var(--showcase-panel); box-shadow: 0 8px 20px var(--showcase-shadow); }
.brand-mark img { width: 28px; height: 28px; object-fit: contain; }
.brand-copy { display: grid; min-width: 0; gap: 3px; }
.brand-copy strong { overflow: hidden; font-size: 14px; text-overflow: ellipsis; white-space: nowrap; }
.brand-copy small, .section-kicker, .route-label, .route-online, .route-node small, .route-visual-bottom, .engine-label, .engine-live, .visual-caption, .hero-eyebrow { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; letter-spacing: .08em; }
.brand-copy small { color: var(--showcase-muted); font-size: 8px; }
.showcase-links { gap: 24px; margin-left: auto; }
.showcase-links a, .showcase-footer a { color: var(--showcase-muted); font-size: 12px; text-decoration: none; transition: color .2s ease; }
.showcase-links a:hover, .showcase-footer a:hover { color: var(--showcase-accent); }
.showcase-actions { gap: 7px; }
.showcase-icon-button { display: grid; width: 36px; height: 36px; place-items: center; border: 1px solid transparent; border-radius: 4px; color: var(--showcase-muted); background: transparent; cursor: pointer; transition: .2s ease; }
.showcase-icon-button:hover, .showcase-icon-button:focus-visible { border-color: var(--showcase-line); color: var(--showcase-accent); background: var(--showcase-panel); outline: none; }
.showcase-login { display: inline-flex; align-items: center; gap: 7px; min-height: 36px; padding: 0 12px; border: 1px solid color-mix(in srgb, var(--showcase-accent) 35%, transparent); border-radius: 4px; color: #f5fffd; background: #101b27; font-size: 12px; font-weight: 700; text-decoration: none; transition: .2s ease; }
.showcase-login:hover, .showcase-login:focus-visible { border-color: var(--showcase-accent); background: color-mix(in srgb, var(--showcase-accent) 25%, #101b27); outline: none; }
.login-avatar { display: grid; width: 20px; height: 20px; place-items: center; border-radius: 3px; color: #061116; background: var(--showcase-accent); font-size: 10px; }

.showcase-hero { display: grid; grid-template-columns: minmax(0, .9fr) minmax(0, 1.1fr); align-items: center; gap: clamp(35px, 7vw, 105px); max-width: 1240px; min-height: min(670px, calc(100vh - 72px)); margin: 0 auto; padding: 74px clamp(18px, 4vw, 64px) 74px; }
.hero-copy { min-width: 0; }
.hero-eyebrow { display: inline-flex; align-items: center; gap: 9px; margin-bottom: 20px; color: var(--showcase-accent); font-size: 10px; font-weight: 700; }
.status-pulse, .route-online i, .tool-status, .engine-live i { display: inline-block; width: 7px; height: 7px; border-radius: 50%; background: var(--showcase-accent); box-shadow: 0 0 12px color-mix(in srgb, var(--showcase-accent) 68%, transparent); animation: pulse-dot 2.4s ease-in-out infinite; }
.hero-copy h1 { max-width: 660px; margin: 0; font-size: clamp(42px, 6.2vw, 78px); line-height: 1.02; letter-spacing: 0; text-wrap: balance; }
.hero-copy h1 span { display: block; margin-top: 10px; color: var(--showcase-accent); text-shadow: 0 0 28px color-mix(in srgb, var(--showcase-accent) 23%, transparent); }
.hero-subtitle { max-width: 570px; margin: 24px 0 0; font-size: clamp(18px, 2vw, 23px); font-weight: 650; line-height: 1.4; }
.hero-description { max-width: 570px; margin: 11px 0 0; color: var(--showcase-muted); font-size: 14px; line-height: 1.8; }
.hero-cta { flex-wrap: wrap; gap: 10px; margin-top: 29px; }
.showcase-primary-button, .showcase-secondary-button { display: inline-flex; align-items: center; justify-content: center; gap: 8px; min-height: 44px; padding: 0 16px; border: 1px solid transparent; border-radius: 4px; font-size: 13px; font-weight: 750; text-decoration: none; transition: transform .2s ease, background .2s ease, border-color .2s ease; }
.showcase-primary-button { color: #061116; background: var(--showcase-accent); box-shadow: 0 8px 22px color-mix(in srgb, var(--showcase-accent) 22%, transparent); }
.showcase-primary-button:hover, .showcase-primary-button:focus-visible { transform: translateY(-2px); background: color-mix(in srgb, var(--showcase-accent) 82%, white); outline: none; }
.showcase-secondary-button { border-color: var(--showcase-line); color: var(--showcase-ink); background: var(--showcase-panel); }
.showcase-secondary-button:hover, .showcase-secondary-button:focus-visible { border-color: var(--showcase-accent); color: var(--showcase-accent); outline: none; }
.hero-metrics { flex-wrap: wrap; gap: 13px 22px; margin-top: 27px; color: var(--showcase-muted); font-size: 10px; }
.hero-metrics span { display: inline-flex; align-items: center; gap: 7px; }
.hero-metrics strong { color: var(--showcase-accent); font: 700 10px ui-monospace, monospace; }

.hero-visual { min-width: 0; }
.terminal-container { position: relative; display: block; width: 100%; }
.route-visual { overflow: hidden; border: 1px solid color-mix(in srgb, var(--showcase-accent) 38%, var(--showcase-line)); border-radius: 4px; background: #0a111b; box-shadow: 0 26px 60px rgba(0, 0, 0, .3), 0 0 0 1px rgba(139, 124, 255, .08) inset; }
.route-visual::before { position: absolute; inset: 0; content: ''; pointer-events: none; opacity: .35; background-image: linear-gradient(rgba(121, 166, 181, .08) 1px, transparent 1px), linear-gradient(90deg, rgba(121, 166, 181, .08) 1px, transparent 1px); background-size: 30px 30px; mask-image: linear-gradient(to bottom, black, transparent 86%); }
.route-visual-top { display: flex; align-items: center; justify-content: space-between; min-height: 45px; padding: 0 18px; border-bottom: 1px solid rgba(143, 174, 205, .16); }
.route-label { color: #91a7bc; font-size: 9px; font-weight: 700; }
.route-online { gap: 7px; color: #79d5bd; font-size: 9px; font-weight: 700; }
.route-graph { position: relative; display: grid; min-height: 330px; padding: 30px 26px; }
.route-node { position: absolute; z-index: 1; display: flex; align-items: center; gap: 10px; min-height: 58px; padding: 10px 13px; border: 1px solid rgba(143, 174, 205, .22); border-radius: 3px; color: #cbd8e6; background: rgba(10, 21, 33, .95); text-align: left; }
.route-node-ingress { top: 25px; left: 26px; width: calc(100% - 52px); }
.route-node-core { top: 128px; left: 50%; width: min(220px, calc(100% - 52px)); transform: translateX(-50%); border-color: color-mix(in srgb, var(--showcase-accent) 55%, rgba(143, 174, 205, .22)); }
.route-node-engine { bottom: 28px; width: calc(50% - 31px); cursor: pointer; transition: border-color .2s ease, background .2s ease, transform .2s ease; }
.route-node-gpt { left: 26px; }
.route-node-claude { right: 26px; }
.route-node-engine:hover, .route-node-engine:focus-visible, .route-node-engine.active { border-color: var(--showcase-accent); background: rgba(69, 230, 209, .1); outline: none; transform: translateY(-3px); }
.route-node-claude:hover, .route-node-claude:focus-visible, .route-node-claude.active { border-color: #e89168; background: rgba(232, 145, 104, .1); }
.route-node-icon, .route-brand { display: grid; width: 32px; height: 32px; flex: 0 0 auto; place-items: center; border: 1px solid rgba(119, 207, 202, .28); border-radius: 3px; color: var(--showcase-accent); background: rgba(69, 230, 209, .08); }
.route-brand { background: #f7fafc; }
.core-icon { border-color: color-mix(in srgb, var(--showcase-violet) 55%, transparent); color: #b3a9ff; background: rgba(139, 124, 255, .1); }
.route-node span:nth-last-child(1) { min-width: 0; }
.route-node > span:not(.route-node-icon):not(.route-brand):not(.core-orbit) { display: grid; gap: 4px; min-width: 0; }
.route-node small { color: #7591a5; font-size: 8px; font-weight: 700; }
.route-node strong { overflow: hidden; color: #eff7ff; font: 650 12px ui-monospace, monospace; text-overflow: ellipsis; white-space: nowrap; }
.route-node-engine > svg { margin-left: auto; color: #7596a7; }
.core-orbit { position: absolute; inset: 8px auto auto 11px; width: 48px; height: 48px; border: 1px solid rgba(139, 124, 255, .26); border-radius: 50%; animation: core-pulse 4s ease-in-out infinite; }
.route-line { position: absolute; z-index: 0; background: rgba(69, 230, 209, .25); }
.route-line span { position: absolute; width: 6px; height: 6px; border-radius: 50%; background: var(--showcase-accent); box-shadow: 0 0 12px var(--showcase-accent); animation: route-flow 3s linear infinite; }
.route-line-main { top: 83px; left: 50%; width: 1px; height: 45px; }
.route-line-main span { left: -3px; top: 0; }
.route-line-gpt, .route-line-claude { top: 186px; width: calc(50% - 42px); height: 1px; }
.route-line-gpt { left: 26px; transform: rotate(28deg); transform-origin: right; }
.route-line-claude { right: 26px; transform: rotate(-28deg); transform-origin: left; }
.route-line-gpt span { right: 0; top: -3px; animation-delay: -.6s; }
.route-line-claude span { left: 0; top: -3px; animation-delay: -1.5s; }
.route-visual-bottom { justify-content: space-between; gap: 12px; min-height: 42px; padding: 0 18px; border-top: 1px solid rgba(143, 174, 205, .16); color: #7695a7; font-size: 9px; }
.route-visual-bottom > span { display: inline-flex; align-items: center; gap: 6px; }
.route-visual-bottom b { color: #71d7b4; font-weight: 700; }
.route-signal { width: 6px; height: 6px; border-radius: 50%; background: #71d7b4; box-shadow: 0 0 9px #71d7b4; }
.visual-caption { display: flex; align-items: center; gap: 7px; margin: 14px 0 0 3px; color: var(--showcase-muted); font-size: 9px; }
.visual-caption span { width: 5px; height: 5px; border-radius: 50%; background: var(--showcase-violet); box-shadow: 0 0 8px var(--showcase-violet); }

.status-strip { display: grid; grid-template-columns: repeat(4, 1fr); max-width: 1112px; margin: 0 auto; border-top: 1px solid var(--showcase-line); border-bottom: 1px solid var(--showcase-line); }
.status-item { gap: 12px; min-height: 72px; padding: 12px 18px; border-right: 1px solid var(--showcase-line); }
.status-item:last-child { border-right: 0; }
.status-item > span:last-child { display: grid; gap: 5px; }
.status-item small { color: var(--showcase-muted); font: 700 8px ui-monospace, monospace; letter-spacing: .08em; }
.status-item strong { font-size: 11px; }
.status-index { color: var(--showcase-accent); font: 700 10px ui-monospace, monospace; }
.status-index-live i { display: block; width: 7px; height: 7px; border-radius: 50%; background: var(--showcase-accent); box-shadow: 0 0 10px var(--showcase-accent); }

.capability-section, .engine-section { max-width: 1112px; margin: 0 auto; padding: 108px 0; }
.section-heading { max-width: 590px; margin-bottom: 38px; }
.section-kicker { display: block; margin-bottom: 13px; color: var(--showcase-accent); font-size: 9px; font-weight: 700; }
.section-heading h2, .tool-copy h2, .engine-intro h2, .showcase-cta h2 { margin: 0; font-size: clamp(28px, 4vw, 46px); line-height: 1.08; letter-spacing: 0; text-wrap: balance; }
.section-heading p, .tool-copy p, .engine-intro p, .showcase-cta p { margin: 13px 0 0; color: var(--showcase-muted); font-size: 14px; line-height: 1.75; }
.capability-rail { display: grid; grid-template-columns: repeat(3, 1fr); border-top: 1px solid var(--showcase-line); border-bottom: 1px solid var(--showcase-line); }
.capability-item { position: relative; min-height: 205px; padding: 25px 27px; border-right: 1px solid var(--showcase-line); }
.capability-item:last-child { border-right: 0; }
.capability-item > svg { color: var(--showcase-accent); }
.capability-index { position: absolute; top: 25px; right: 27px; color: var(--showcase-muted); font: 700 10px ui-monospace, monospace; }
.capability-item h3 { margin: 20px 0 8px; font-size: 16px; }
.capability-item p { margin: 0; color: var(--showcase-muted); font-size: 13px; line-height: 1.7; }

.tool-section { display: grid; grid-template-columns: minmax(0, .72fr) minmax(0, 1.28fr); align-items: center; gap: 56px; max-width: 1112px; margin: 0 auto; padding: 0 0 108px; }
.tool-copy { max-width: 395px; }
.tool-rail { position: relative; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
.tool-rail-line { position: absolute; top: 15px; bottom: 15px; left: 50%; width: 1px; background: linear-gradient(to bottom, transparent, color-mix(in srgb, var(--showcase-accent) 70%, transparent), transparent); transform: translateX(-50%); }
.tool-rail-line span { position: absolute; top: 18%; left: -3px; width: 7px; height: 7px; border-radius: 50%; background: var(--showcase-accent); box-shadow: 0 0 12px var(--showcase-accent); animation: rail-flow 5s linear infinite; }
.tool-item { position: relative; z-index: 1; display: flex; align-items: center; gap: 11px; min-width: 0; min-height: 76px; padding: 13px 14px; border: 1px solid var(--showcase-line); border-radius: 3px; background: var(--showcase-panel); transition: border-color .2s ease, transform .2s ease; }
.tool-item:hover, .tool-item:focus-within { border-color: var(--showcase-accent); transform: translateY(-2px); }
.tool-icon-frame { display: grid; width: 43px; height: 43px; flex: 0 0 auto; place-items: center; border: 1px solid var(--showcase-line); border-radius: 3px; background: #f6fafc; }
.tool-item > span:nth-child(2) { display: grid; min-width: 0; gap: 5px; }
.tool-item strong { overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.tool-item small { color: var(--showcase-muted); font: 700 8px ui-monospace, monospace; letter-spacing: .07em; }
.tool-status { width: 5px; height: 5px; margin-left: auto; animation-delay: -1s; }

.engine-section { display: grid; grid-template-columns: minmax(0, .72fr) minmax(0, 1.28fr); gap: 60px; padding-top: 0; }
.engine-intro { align-self: center; }
.inline-link { display: inline-flex; align-items: center; gap: 7px; margin-top: 22px; color: var(--showcase-accent); font-size: 13px; font-weight: 750; text-decoration: none; }
.inline-link:hover { color: color-mix(in srgb, var(--showcase-accent) 75%, white); }
.engine-cards { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.engine-card { min-height: 225px; padding: 21px; border: 1px solid var(--showcase-line); border-radius: 3px; background: var(--showcase-panel); box-shadow: 0 14px 36px var(--showcase-shadow); transition: transform .2s ease, border-color .2s ease; }
.engine-card:hover { transform: translateY(-4px); border-color: var(--showcase-accent); }
.engine-card-head { gap: 9px; }
.engine-large-logo { display: grid; width: 42px; height: 42px; place-items: center; border: 1px solid var(--showcase-line); border-radius: 3px; background: #f6fafc; }
.engine-label { color: var(--showcase-muted); font-size: 8px; }
.engine-live { gap: 5px; margin-left: auto; color: var(--showcase-accent); font-size: 8px; }
.engine-live i { width: 5px; height: 5px; }
.engine-card h3 { margin: 27px 0 8px; font-size: 25px; }
.engine-card p { min-height: 50px; margin: 0; color: var(--showcase-muted); font-size: 12px; line-height: 1.7; }
.engine-card-foot { justify-content: space-between; margin-top: 25px; padding-top: 13px; border-top: 1px solid var(--showcase-line); color: var(--showcase-muted); font: 700 8px ui-monospace, monospace; letter-spacing: .08em; }
.engine-card-foot svg { color: var(--showcase-accent); }

.showcase-cta { display: flex; align-items: center; justify-content: space-between; gap: 28px; max-width: 1112px; margin: 0 auto 88px; padding: 35px 42px; border: 1px solid var(--showcase-line); border-radius: 3px; background: var(--showcase-bg-soft); }
.showcase-cta h2 { font-size: clamp(24px, 3vw, 36px); }
.showcase-cta p { max-width: 550px; }
.showcase-footer { display: flex; align-items: center; justify-content: space-between; gap: 18px; max-width: 1112px; margin: 0 auto; padding: 22px 0 32px; border-top: 1px solid var(--showcase-line); color: var(--showcase-muted); font-size: 11px; }
.showcase-footer > div { gap: 20px; }

@keyframes pulse-dot { 0%, 100% { opacity: 1; } 50% { opacity: .42; } }
@keyframes route-flow { from { top: 0; } to { top: calc(100% - 6px); } }
@keyframes rail-flow { from { top: 18%; } to { top: 82%; } }
@keyframes core-pulse { 0%, 100% { opacity: .3; transform: scale(.92); } 50% { opacity: .85; transform: scale(1.04); } }

@media (max-width: 980px) {
  .showcase-links { display: none; }
  .showcase-hero { grid-template-columns: 1fr; min-height: auto; padding-top: 67px; }
  .hero-copy { max-width: 720px; }
  .hero-visual { width: min(100%, 700px); margin: 0 auto; }
  .capability-section, .tool-section, .engine-section, .showcase-footer { margin-right: 18px; margin-left: 18px; }
  .tool-section { grid-template-columns: 1fr; gap: 35px; }
  .tool-copy { max-width: 600px; }
  .engine-section { gap: 38px; }
}

@media (max-width: 680px) {
  .showcase-header { padding: 11px 14px; }
  .showcase-nav { gap: 10px; }
  .brand-mark { width: 36px; height: 36px; }
  .brand-mark img { width: 25px; height: 25px; }
  .brand-copy { max-width: min(36vw, 170px); }
  .brand-copy small, .showcase-actions > :deep(.locale-switcher) { display: none; }
  .showcase-actions { gap: 3px; }
  .showcase-icon-button { width: 33px; height: 33px; }
  .showcase-login { min-height: 33px; padding: 0 9px; font-size: 11px; }
  .showcase-login svg { display: none; }
  .showcase-hero { gap: 38px; padding: 53px 18px 56px; }
  .hero-copy h1 { width: 100%; max-width: 100%; font-size: clamp(38px, 12.5vw, 54px); overflow-wrap: anywhere; }
  .hero-subtitle, .hero-description { max-width: 100%; overflow-wrap: anywhere; }
  .hero-cta { align-items: stretch; flex-direction: column; }
  .showcase-primary-button, .showcase-secondary-button { width: 100%; }
  .hero-metrics { gap: 8px 13px; font-size: 9px; }
  .route-graph { min-height: 360px; padding: 24px 13px; }
  .route-node-ingress { left: 13px; width: calc(100% - 26px); }
  .route-node-core { top: 132px; width: calc(100% - 26px); }
  .route-node-engine { bottom: 24px; width: calc(50% - 20px); min-height: 67px; padding: 8px; }
  .route-node-gpt { left: 13px; }
  .route-node-claude { right: 13px; }
  .route-brand, .route-node-icon { width: 28px; height: 28px; }
  .route-node strong { font-size: 10px; }
  .route-node small { font-size: 7px; }
  .route-line-gpt, .route-line-claude { width: calc(50% - 25px); }
  .route-visual-bottom { flex-wrap: wrap; gap: 6px 12px; min-height: 50px; padding: 8px 12px; font-size: 8px; }
  .status-strip { grid-template-columns: repeat(2, 1fr); margin: 0 18px; }
  .status-item { min-height: 64px; padding: 11px 10px; border-bottom: 1px solid var(--showcase-line); }
  .status-item:nth-child(2n) { border-right: 0; }
  .status-item:nth-child(3), .status-item:nth-child(4) { border-bottom: 0; }
  .capability-section, .tool-section, .engine-section { padding: 75px 0; }
  .tool-section { padding-top: 0; }
  .tool-rail { grid-template-columns: 1fr; gap: 8px; }
  .tool-rail-line { top: 14px; bottom: 14px; left: 4px; }
  .tool-item { min-height: 66px; padding: 10px 11px 10px 17px; }
  .capability-rail, .engine-section, .engine-cards { grid-template-columns: 1fr; }
  .capability-item { min-height: 175px; border-right: 0; border-bottom: 1px solid var(--showcase-line); }
  .capability-item:last-child { border-bottom: 0; }
  .engine-section { gap: 33px; padding-top: 0; }
  .showcase-cta { align-items: stretch; flex-direction: column; margin: 0 18px 65px; padding: 27px 20px; }
  .showcase-footer { align-items: flex-start; flex-direction: column; padding-bottom: 24px; }
}

@media (prefers-reduced-motion: reduce) {
  .status-pulse, .route-online i, .tool-status, .engine-live i, .route-line span, .tool-rail-line span, .core-orbit { animation: none; }
  .showcase-primary-button, .showcase-login, .route-node-engine, .tool-item, .engine-card { transition: none; }
}
</style>

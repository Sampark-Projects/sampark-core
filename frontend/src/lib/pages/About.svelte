<script>
  import { Scale } from '@lucide/svelte';
  import Button from '../components/Button.svelte';
  import { navigate } from '../router.js';
  import { versionLabel } from '../version.js';
  import { t } from '../stores/i18n.svelte.js';

  let { version = versionLabel } = $props();
</script>

<div class="min-h-screen" style="background-color: var(--ds-surface);">
  <div class="max-w-4xl mx-auto px-6 py-12">

    <!-- Logo and Title -->
    <div class="text-center mb-12">
      <div class="flex justify-center mb-6">
        <img src="spark-logo.png" alt="Spark" class="h-24 w-24" />
      </div>
      <h1 class="text-4xl font-bold mb-4" style="color: var(--ds-text);">
        {t('about.title')}
      </h1>
      <p class="text-xl" style="color: var(--ds-text-subtle);">
        {t('about.subtitle')}
      </p>
    </div>

    <!-- Main Content -->
    <div class="space-y-12">
      <!-- About Section -->
      <section>
        <h2 class="text-2xl font-semibold mb-6" style="color: var(--ds-text);">{t('about.whatIs')}</h2>
        <div class="prose max-w-none" style="color: var(--ds-text);">
          <p class="text-lg leading-relaxed mb-4">
            {t('about.description')}
          </p>
        </div>
      </section>

      <!-- Notable Libraries -->
      <section>
        <h2 class="text-2xl font-semibold mb-6" style="color: var(--ds-text);">{t('about.builtWith')}</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {#each [
            { name: 'Svelte', desc: t('about.libSvelte') },
            { name: 'Tailwind CSS', desc: t('about.libTailwind') },
            { name: 'Milkdown', desc: t('about.libMilkdown') },
            { name: 'Excalidraw', desc: t('about.libExcalidraw') },
            { name: 'Melt UI', desc: t('about.libMeltUI') },
            { name: 'Extism', desc: t('about.libExtism') },
            { name: 'Charmbracelet', desc: t('about.libCharmbracelet') },
            { name: 'Svelte Flow', desc: t('about.libSvelteFlow') },
          ] as lib}
            <div class="p-4 rounded border" style="border-color: var(--ds-border); background-color: var(--ds-surface-raised);">
              <div class="font-semibold mb-1" style="color: var(--ds-text);">{lib.name}</div>
              <div class="text-sm" style="color: var(--ds-text-subtle);">{lib.desc}</div>
            </div>
          {/each}
        </div>
      </section>

      <!-- Version Info -->
      <section class="text-center py-8">
        <div class="p-8 rounded border" style="border-color: var(--ds-border); background-color: var(--ds-surface-raised);">
          <h2 class="text-xl font-semibold mb-4" style="color: var(--ds-text);">{t('about.versionInfo')}</h2>
          <div class="space-y-2">
            <div>
              <span class="font-medium" style="color: var(--ds-text);">{t('about.version')}:</span>
              <span class="ml-2" style="color: var(--ds-text-subtle);">{version}</span>
            </div>
            <div>
              <span style="color: var(--ds-text-subtle);">{t('about.developedBy')} </span>
              <a href="https://realigned.io" target="_blank" rel="noopener noreferrer" class="font-medium hover:underline" style="color: var(--ds-interactive);">Realigned Technologies Ltd</a>
            </div>
          </div>
        </div>
      </section>

      <!-- Footer Actions -->
      <section class="text-center">
        <div class="flex justify-center gap-4">
          <Button
            onclick={() => navigate('/licenses')}
            variant="default"
            size="medium"
            icon={Scale}
          >
            Licenses
          </Button>
        </div>
      </section>
    </div>
  </div>
</div>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhLinkSimple, PhCode, PhBrowser } from '@phosphor-icons/vue';

interface Props {
  modelValue: 'url' | 'script' | 'xpath';
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: 'url' | 'script' | 'xpath'];
}>();

const { t } = useI18n();

const modes = [
  {
    key: 'url' as const,
    label: t('rssUrl'),
    icon: PhLinkSimple,
    description: t('rssUrlDescription'),
  },
  {
    key: 'script' as const,
    label: t('customScript'),
    icon: PhCode,
    description: t('customScriptDescription'),
  },
  {
    key: 'xpath' as const,
    label: t('xpath'),
    icon: PhBrowser,
    description: t('xpathDescription'),
  },
];

function selectMode(mode: 'url' | 'script' | 'xpath') {
  emit('update:modelValue', mode);
}
</script>

<template>
  <div class="mb-4 sm:mb-6">
    <!-- Primary RSS URL mode -->
    <div class="mb-4">
      <div
        :class="[
          'p-4 rounded-lg border-2 transition-all duration-200',
          props.modelValue === 'url'
            ? 'border-accent bg-accent/5 shadow-sm'
            : 'border-border bg-bg-tertiary',
        ]"
      >
        <button
          type="button"
          class="w-full flex items-center gap-3 text-left"
          @click="selectMode('url')"
        >
          <div
            :class="[
              'w-12 h-12 rounded-full flex items-center justify-center transition-colors',
              props.modelValue === 'url'
                ? 'bg-accent text-white'
                : 'bg-bg-secondary text-text-secondary',
            ]"
          >
            <PhLinkSimple :size="20" />
          </div>
          <div class="flex-1">
            <div class="font-semibold text-base sm:text-lg">{{ t('rssUrl') }}</div>
            <div class="text-sm text-text-secondary">{{ t('rssUrlDescription') }}</div>
          </div>
          <div
            v-if="props.modelValue === 'url'"
            class="w-6 h-6 rounded-full bg-accent text-white flex items-center justify-center text-sm font-bold"
          >
            ✓
          </div>
        </button>
      </div>
    </div>

    <!-- Advanced options as subtle links -->
    <div class="flex items-center justify-center gap-4 text-xs">
      <span class="text-text-secondary">{{ t('orTry') }}</span>
      <button
        type="button"
        :class="[
          'hover:text-accent transition-colors',
          props.modelValue === 'script' ? 'text-accent font-medium' : 'text-text-secondary',
        ]"
        @click="selectMode('script')"
      >
        <PhCode :size="14" class="inline mr-1" />
        {{ t('customScript') }}
      </button>
      <span class="text-text-secondary">•</span>
      <button
        type="button"
        :class="[
          'hover:text-accent transition-colors',
          props.modelValue === 'xpath' ? 'text-accent font-medium' : 'text-text-secondary',
        ]"
        @click="selectMode('xpath')"
      >
        <PhBrowser :size="14" class="inline mr-1" />
        {{ t('xpath') }}
      </button>
    </div>

    <!-- Current mode indicator (only shown for advanced modes) -->
    <div
      v-if="props.modelValue !== 'url'"
      class="mt-4 p-3 bg-accent/10 border border-accent/30 rounded-md text-xs"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 bg-accent rounded-full animate-pulse"></div>
          <span class="text-accent font-medium">{{ t('currentMode') }}:</span>
          <strong>{{ modes.find((m) => m.key === props.modelValue)?.label }}</strong>
        </div>
        <button
          type="button"
          class="text-accent hover:text-accent-hover underline"
          @click="selectMode('url')"
        >
          {{ t('backToRss') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

/* Additional custom animations */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-slide-in {
  animation: slideIn 0.3s ease-out;
}
</style>

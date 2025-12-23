<script setup lang="ts">
/* eslint-disable vue/no-v-html */
import { PhSpinnerGap, PhArticle, PhArrowClockwise } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface Props {
  articleContent: string;
  isTranslatingContent: boolean;
  hasMediaContent?: boolean; // Whether article has audio/video content
  isLoadingContent?: boolean; // Whether content is currently loading
}

const props = withDefaults(defineProps<Props>(), {
  hasMediaContent: false,
  isLoadingContent: false,
});

// Emits
const emit = defineEmits<{
  retryLoad: [];
}>();
</script>

<template>
  <!-- Content display with inline translations -->
  <div v-if="articleContent">
    <div
      class="prose prose-sm sm:prose-lg max-w-none text-text-primary prose-content"
      v-html="articleContent"
    ></div>
    <!-- Translation loading indicator -->
    <div v-if="isTranslatingContent" class="flex items-center gap-2 mt-4 text-text-secondary">
      <PhSpinnerGap :size="16" class="animate-spin" />
      <span class="text-sm">{{ t('translatingContent') }}</span>
    </div>
  </div>

  <!-- No content available with retry option -->
  <div v-else-if="!hasMediaContent" class="text-center text-text-secondary py-6 sm:py-8">
    <PhArticle :size="48" class="mb-2 sm:mb-3 opacity-50 mx-auto sm:w-16 sm:h-16" />
    <p class="text-sm sm:text-base mb-4">{{ t('noContentAvailable') }}</p>
    <button
      v-if="!props.isLoadingContent"
      class="btn-secondary-compact flex items-center gap-1.5 mx-auto"
      @click="emit('retryLoad')"
    >
      <PhArrowClockwise :size="12" />
      <span class="text-xs">{{ t('retrySummary') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

interface Props {
  modelValue: string;
  mode: 'add' | 'edit';
  isInvalid?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  isInvalid: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const { t } = useI18n();
</script>

<template>
  <div class="mb-3 sm:mb-4">
    <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary"
      >{{ t('rssUrl') }} <span v-if="props.mode === 'add'" class="text-red-500">*</span></label
    >
    <input
      :value="props.modelValue"
      type="text"
      :placeholder="t('rssUrlPlaceholder')"
      :class="['input-field', props.mode === 'add' && props.isInvalid ? 'border-red-500' : '']"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
}
</style>

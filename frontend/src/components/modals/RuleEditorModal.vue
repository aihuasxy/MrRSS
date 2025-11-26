<script setup>
import { ref, computed, onMounted, watch } from 'vue';
import { store } from '../../store.js';
import { 
    PhLightning, PhX, PhPlus, PhTrash, PhFunnel, PhProhibit, PhListChecks
} from "@phosphor-icons/vue";

const props = defineProps({
    show: { type: Boolean, default: false },
    rule: { type: Object, default: null }
});

const emit = defineEmits(['close', 'save']);

// Form data
const ruleName = ref('');
const conditions = ref([]);
const actions = ref([]);

// Field options (reusing from ArticleFilterModal)
const fieldOptions = [
    { value: 'feed_name', labelKey: 'feedName', multiSelect: true },
    { value: 'feed_category', labelKey: 'feedCategory', multiSelect: true },
    { value: 'article_title', labelKey: 'articleTitle', multiSelect: false },
    { value: 'published_after', labelKey: 'publishedAfter', multiSelect: false },
    { value: 'published_before', labelKey: 'publishedBefore', multiSelect: false },
    { value: 'is_read', labelKey: 'readStatus', multiSelect: false, booleanField: true },
    { value: 'is_favorite', labelKey: 'favoriteStatus', multiSelect: false, booleanField: true },
    { value: 'is_hidden', labelKey: 'hiddenStatus', multiSelect: false, booleanField: true }
];

// Operator options for article title
const textOperatorOptions = [
    { value: 'contains', labelKey: 'contains' },
    { value: 'exact', labelKey: 'exactMatch' }
];

// Boolean value options
const booleanOptions = [
    { value: 'true', labelKey: 'yes' },
    { value: 'false', labelKey: 'no' }
];

// Logic options
const logicOptions = [
    { value: 'and', labelKey: 'and' },
    { value: 'or', labelKey: 'or' }
];

// Action options
const actionOptions = [
    { value: 'favorite', labelKey: 'actionFavorite' },
    { value: 'unfavorite', labelKey: 'actionUnfavorite' },
    { value: 'hide', labelKey: 'actionHide' },
    { value: 'unhide', labelKey: 'actionUnhide' },
    { value: 'mark_read', labelKey: 'actionMarkRead' },
    { value: 'mark_unread', labelKey: 'actionMarkUnread' }
];

// Feed names for multi-select
const feedNames = computed(() => {
    return store.feeds.map(f => f.title);
});

// Feed categories for multi-select
const feedCategories = computed(() => {
    const categories = new Set();
    store.feeds.forEach(f => {
        if (f.category) {
            categories.add(f.category);
        }
    });
    return Array.from(categories);
});

// Initialize form when rule changes
watch(() => props.rule, (newRule) => {
    if (newRule) {
        ruleName.value = newRule.name || '';
        conditions.value = newRule.conditions ? JSON.parse(JSON.stringify(newRule.conditions)) : [];
        actions.value = newRule.actions ? [...newRule.actions] : [];
    } else {
        ruleName.value = '';
        conditions.value = [];
        actions.value = [];
    }
}, { immediate: true });

// Condition helpers
function addCondition() {
    conditions.value.push({
        id: Date.now(),
        logic: conditions.value.length > 0 ? 'and' : null,
        negate: false,
        field: 'article_title',
        operator: 'contains',
        value: '',
        values: []
    });
}

function removeCondition(index) {
    conditions.value.splice(index, 1);
    if (conditions.value.length > 0 && index === 0) {
        conditions.value[0].logic = null;
    }
}

function isDateField(field) {
    return field === 'published_after' || field === 'published_before';
}

function isMultiSelectField(field) {
    return field === 'feed_name' || field === 'feed_category';
}

function isBooleanField(field) {
    return field === 'is_read' || field === 'is_favorite';
}

function needsOperator(field) {
    return field === 'article_title';
}

function onFieldChange(index) {
    const condition = conditions.value[index];
    if (isDateField(condition.field)) {
        condition.operator = null;
        condition.value = '';
        condition.values = [];
    } else if (isMultiSelectField(condition.field)) {
        condition.operator = 'contains';
        condition.value = '';
        condition.values = [];
    } else if (isBooleanField(condition.field)) {
        condition.operator = null;
        condition.value = 'true';
        condition.values = [];
    } else {
        condition.operator = 'contains';
        condition.value = '';
        condition.values = [];
    }
}

function toggleNegate(index) {
    conditions.value[index].negate = !conditions.value[index].negate;
}

// Dropdown for multi-select
const openDropdownIndex = ref(null);

function toggleDropdown(index) {
    if (openDropdownIndex.value === index) {
        openDropdownIndex.value = null;
    } else {
        openDropdownIndex.value = index;
    }
}

function toggleMultiSelectValue(index, val) {
    const condition = conditions.value[index];
    const idx = condition.values.indexOf(val);
    if (idx > -1) {
        condition.values.splice(idx, 1);
    } else {
        condition.values.push(val);
    }
}

function getMultiSelectDisplayText(condition, labelKey) {
    if (!condition.values || condition.values.length === 0) {
        return store.i18n.t(labelKey);
    }
    
    if (condition.values.length === 1) {
        return condition.values[0];
    }
    
    // Show first item and total count
    // For Chinese: "xxx等N个" means "xxx and N items in total"
    // For English: "xxx and N more" means N additional items
    const firstItem = condition.values[0];
    const totalCount = condition.values.length;
    const remaining = totalCount - 1;
    
    // Use different count based on locale
    const locale = store.i18n.locale.value;
    if (locale === 'zh') {
        return `${firstItem} ${store.i18n.t('andNMore', { count: totalCount })}`;
    }
    return `${firstItem} ${store.i18n.t('andNMore', { count: remaining })}`;
}

// Action helpers
function addAction() {
    // Find first action that isn't already selected
    const selectedActions = new Set(actions.value);
    const available = actionOptions.find(opt => !selectedActions.has(opt.value));
    if (available) {
        actions.value.push(available.value);
    }
}

function removeAction(index) {
    actions.value.splice(index, 1);
}

function updateAction(index, value) {
    actions.value[index] = value;
}

function getAvailableActions(currentValue) {
    const selectedActions = new Set(actions.value);
    return actionOptions.filter(opt => !selectedActions.has(opt.value) || opt.value === currentValue);
}

// Form validation
const isValid = computed(() => {
    return actions.value.length > 0;
});

// Save handler
function handleSave() {
    if (!isValid.value) {
        window.showToast(store.i18n.t('noActionsSelected'), 'warning');
        return;
    }
    
    const rule = {
        id: props.rule ? props.rule.id : Date.now(),
        name: ruleName.value || store.i18n.t('rules'),
        enabled: props.rule ? props.rule.enabled : true,
        conditions: conditions.value.filter(c => {
            if (isMultiSelectField(c.field)) {
                return c.values && c.values.length > 0;
            }
            return c.value !== '';
        }),
        actions: [...actions.value]
    };
    
    emit('save', rule);
}

function handleClose() {
    openDropdownIndex.value = null;
    emit('close');
}
</script>

<template>
    <div v-if="show" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <div class="bg-bg-primary w-full max-w-2xl max-h-[90vh] flex flex-col rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in">
            <!-- Header -->
            <div class="p-4 sm:p-5 border-b border-border flex justify-between items-center shrink-0">
                <h3 class="text-lg font-semibold m-0 flex items-center gap-2">
                    <PhLightning :size="20" />
                    {{ rule ? store.i18n.t('editRule') : store.i18n.t('addRule') }}
                </h3>
                <span @click="handleClose" class="text-2xl cursor-pointer text-text-secondary hover:text-text-primary">&times;</span>
            </div>
            
            <!-- Content -->
            <div class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
                <!-- Rule Name -->
                <div class="space-y-2">
                    <label class="block text-sm font-medium">{{ store.i18n.t('ruleName') }}</label>
                    <input 
                        type="text" 
                        v-model="ruleName" 
                        :placeholder="store.i18n.t('ruleNamePlaceholder')"
                        class="input-field w-full"
                    >
                </div>
                
                <!-- Conditions Section -->
                <div class="space-y-3">
                    <div class="flex items-center justify-between">
                        <label class="flex items-center gap-2 text-sm font-medium">
                            <PhFunnel :size="16" />
                            {{ store.i18n.t('ruleCondition') }}
                        </label>
                    </div>
                    
                    <!-- Empty state -->
                    <div v-if="conditions.length === 0" class="text-center text-text-secondary py-4 bg-bg-secondary rounded-lg border border-border">
                        <p class="text-sm">{{ store.i18n.t('conditionAlways') }}</p>
                    </div>
                    
                    <!-- Condition list -->
                    <div v-else class="space-y-3">
                        <div v-for="(condition, index) in conditions" :key="condition.id">
                            <!-- Logic connector -->
                            <div v-if="index > 0" class="flex items-center justify-center my-3">
                                <div class="flex-1 h-px bg-border"></div>
                                <div class="logic-connector mx-3">
                                    <button 
                                        v-for="opt in logicOptions" 
                                        :key="opt.value"
                                        @click="condition.logic = opt.value"
                                        :class="['logic-btn', condition.logic === opt.value ? 'active' : '']">
                                        {{ store.i18n.t(opt.labelKey) }}
                                    </button>
                                </div>
                                <div class="flex-1 h-px bg-border"></div>
                            </div>
                            
                            <!-- Condition card -->
                            <div class="condition-row bg-bg-secondary border border-border rounded-lg p-3">
                                <div class="flex flex-wrap gap-2 items-end">
                                    <!-- NOT toggle button -->
                                    <div class="flex-shrink-0">
                                        <label class="block text-xs text-text-secondary mb-1">&nbsp;</label>
                                        <button 
                                            @click="toggleNegate(index)" 
                                            :class="['not-btn', condition.negate ? 'active' : '']"
                                            :title="store.i18n.t('not')">
                                            <PhProhibit :size="16" />
                                            <span class="text-xs font-medium">{{ store.i18n.t('not') }}</span>
                                        </button>
                                    </div>
                                    
                                    <!-- Field selector -->
                                    <div class="flex-1 min-w-[130px]">
                                        <label class="block text-xs text-text-secondary mb-1">{{ store.i18n.t('filterField') }}</label>
                                        <select v-model="condition.field" @change="onFieldChange(index)" class="select-field w-full">
                                            <option v-for="opt in fieldOptions" :key="opt.value" :value="opt.value">
                                                {{ store.i18n.t(opt.labelKey) }}
                                            </option>
                                        </select>
                                    </div>
                                    
                                    <!-- Operator selector (only for article_title) -->
                                    <div v-if="needsOperator(condition.field)" class="w-28">
                                        <label class="block text-xs text-text-secondary mb-1">{{ store.i18n.t('filterOperator') }}</label>
                                        <select v-model="condition.operator" class="select-field w-full">
                                            <option v-for="opt in textOperatorOptions" :key="opt.value" :value="opt.value">
                                                {{ store.i18n.t(opt.labelKey) }}
                                            </option>
                                        </select>
                                    </div>
                                    
                                    <!-- Value input -->
                                    <div class="flex-1 min-w-[140px]">
                                        <label class="block text-xs text-text-secondary mb-1">{{ store.i18n.t('filterValue') }}</label>
                                        
                                        <!-- Date input -->
                                        <input v-if="isDateField(condition.field)" 
                                               type="date" 
                                               v-model="condition.value" 
                                               class="date-field w-full">
                                        
                                        <!-- Boolean select -->
                                        <select v-else-if="isBooleanField(condition.field)"
                                                v-model="condition.value"
                                                class="select-field w-full">
                                            <option v-for="opt in booleanOptions" :key="opt.value" :value="opt.value">
                                                {{ store.i18n.t(opt.labelKey) }}
                                            </option>
                                        </select>
                                        
                                        <!-- Multi-select dropdown for feed name -->
                                        <div v-else-if="condition.field === 'feed_name'" class="dropdown-container">
                                            <button type="button" 
                                                    @click="toggleDropdown(index)"
                                                    class="dropdown-trigger">
                                                <span class="dropdown-text truncate">{{ getMultiSelectDisplayText(condition, 'feedName') }}</span>
                                                <span class="dropdown-arrow">▼</span>
                                            </button>
                                            <div v-if="openDropdownIndex === index" 
                                                 class="dropdown-menu dropdown-down">
                                                <div v-for="name in feedNames" :key="name" 
                                                     @click.stop="toggleMultiSelectValue(index, name)"
                                                     :class="['dropdown-option', condition.values.includes(name) ? 'selected' : '']">
                                                    <input type="checkbox" 
                                                           :checked="condition.values.includes(name)" 
                                                           class="checkbox-input"
                                                           tabindex="-1">
                                                    <span class="truncate">{{ name }}</span>
                                                </div>
                                                <div v-if="feedNames.length === 0" class="text-text-secondary text-sm p-2">
                                                    {{ store.i18n.t('noArticles') }}
                                                </div>
                                            </div>
                                        </div>
                                        
                                        <!-- Multi-select dropdown for category -->
                                        <div v-else-if="condition.field === 'feed_category'" class="dropdown-container">
                                            <button type="button" 
                                                    @click="toggleDropdown(index)"
                                                    class="dropdown-trigger">
                                                <span class="dropdown-text truncate">{{ getMultiSelectDisplayText(condition, 'feedCategory') }}</span>
                                                <span class="dropdown-arrow">▼</span>
                                            </button>
                                            <div v-if="openDropdownIndex === index" 
                                                 class="dropdown-menu dropdown-down">
                                                <div v-for="cat in feedCategories" :key="cat" 
                                                     @click.stop="toggleMultiSelectValue(index, cat)"
                                                     :class="['dropdown-option', condition.values.includes(cat) ? 'selected' : '']">
                                                    <input type="checkbox" 
                                                           :checked="condition.values.includes(cat)" 
                                                           class="checkbox-input"
                                                           tabindex="-1">
                                                    <span class="truncate">{{ cat }}</span>
                                                </div>
                                                <div v-if="feedCategories.length === 0" class="text-text-secondary text-sm p-2">
                                                    {{ store.i18n.t('noArticles') }}
                                                </div>
                                            </div>
                                        </div>
                                        
                                        <!-- Regular text input -->
                                        <input v-else 
                                               type="text" 
                                               v-model="condition.value" 
                                               class="input-field w-full"
                                               :placeholder="store.i18n.t('filterValue')">
                                    </div>
                                    
                                    <!-- Remove button -->
                                    <div class="flex-shrink-0">
                                        <label class="block text-xs text-text-secondary mb-1">&nbsp;</label>
                                        <button @click="removeCondition(index)" class="btn-danger-icon" :title="store.i18n.t('removeCondition')">
                                            <PhTrash :size="18" />
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <!-- Add condition button -->
                    <button @click="addCondition" class="btn-secondary w-full flex items-center justify-center gap-2">
                        <PhPlus :size="16" />
                        {{ store.i18n.t('addCondition') }}
                    </button>
                </div>
                
                <!-- Actions Section -->
                <div class="space-y-3">
                    <div class="flex items-center justify-between">
                        <label class="flex items-center gap-2 text-sm font-medium">
                            <PhListChecks :size="16" />
                            {{ store.i18n.t('ruleActions') }}
                        </label>
                    </div>
                    
                    <!-- Empty state -->
                    <div v-if="actions.length === 0" class="text-center text-text-secondary py-4 bg-bg-secondary rounded-lg border border-border">
                        <p class="text-sm">{{ store.i18n.t('noActionsSelected') }}</p>
                    </div>
                    
                    <!-- Action list -->
                    <div v-else class="space-y-2">
                        <div v-for="(action, index) in actions" :key="index" class="action-row">
                            <span class="text-xs text-text-secondary">{{ index + 1 }}.</span>
                            <select :value="action" @change="updateAction(index, $event.target.value)" class="select-field flex-1">
                                <option v-for="opt in getAvailableActions(action)" :key="opt.value" :value="opt.value">
                                    {{ store.i18n.t(opt.labelKey) }}
                                </option>
                            </select>
                            <button @click="removeAction(index)" class="btn-danger-icon" :title="store.i18n.t('removeAction')">
                                <PhTrash :size="16" />
                            </button>
                        </div>
                    </div>
                    
                    <!-- Add action button -->
                    <button 
                        @click="addAction" 
                        class="btn-secondary w-full flex items-center justify-center gap-2"
                        :disabled="actions.length >= actionOptions.length"
                    >
                        <PhPlus :size="16" />
                        {{ store.i18n.t('addAction') }}
                    </button>
                </div>
            </div>
            
            <!-- Footer -->
            <div class="p-4 sm:p-5 border-t border-border bg-bg-secondary flex justify-end gap-3 shrink-0">
                <button @click="handleClose" class="btn-secondary">
                    {{ store.i18n.t('cancel') }}
                </button>
                <button @click="handleSave" class="btn-primary" :disabled="!isValid">
                    {{ store.i18n.t('saveChanges') }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.input-field {
    @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors;
}
.select-field {
    @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors cursor-pointer;
}
.date-field {
    @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors cursor-pointer;
    color-scheme: light dark;
}
.btn-primary {
    @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}
.btn-secondary {
    @apply bg-bg-tertiary text-text-primary border border-border px-4 py-2.5 rounded-lg cursor-pointer font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}
.btn-danger-icon {
    @apply p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors cursor-pointer;
}

/* Logic connector styling */
.logic-connector {
    @apply flex items-center gap-1 bg-bg-tertiary rounded-full p-1;
}
.logic-btn {
    @apply px-3 py-1 text-xs font-bold rounded-full transition-all cursor-pointer;
    @apply text-text-secondary bg-transparent;
}
.logic-btn:hover {
    @apply text-text-primary bg-bg-secondary;
}
.logic-btn.active {
    @apply text-white bg-accent;
}

/* NOT button styling */
.not-btn {
    @apply flex items-center gap-1 px-2 py-2 rounded-md border transition-all cursor-pointer;
    @apply text-text-secondary bg-bg-primary border-border;
}
.not-btn:hover {
    @apply border-red-400 text-red-500;
}
.not-btn.active {
    @apply bg-red-500/10 border-red-500 text-red-500;
}

/* Dropdown multi-select styling */
.dropdown-container {
    @apply relative;
}
.dropdown-trigger {
    @apply w-full p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm;
    @apply flex items-center justify-between cursor-pointer hover:border-accent transition-colors;
}
.dropdown-text {
    @apply flex-1 text-left;
}
.dropdown-arrow {
    @apply text-text-secondary text-xs ml-2;
}
.dropdown-menu {
    @apply absolute left-0 right-0 border border-border rounded-md bg-bg-primary;
    @apply max-h-40 overflow-y-auto z-50 shadow-lg;
}
.dropdown-menu.dropdown-down {
    @apply top-full mt-1;
}
.dropdown-option {
    @apply flex items-center gap-2 px-3 py-2 cursor-pointer text-sm text-text-primary hover:bg-bg-tertiary;
}
.dropdown-option.selected {
    background-color: rgba(59, 130, 246, 0.1);
}
.checkbox-input {
    @apply w-4 h-4 accent-accent cursor-pointer;
}

/* Action row */
.action-row {
    @apply flex items-center gap-2 p-2 bg-bg-secondary border border-border rounded-lg;
}

.animate-fade-in {
    animation: modalFadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes modalFadeIn {
    from { transform: translateY(-20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}
</style>

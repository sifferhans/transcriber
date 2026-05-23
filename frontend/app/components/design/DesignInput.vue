<script setup lang="ts">
// Mirrors Phoenix's CoreComponents.input/1 — a unified field component that
// branches on `type` (text-like, textarea, select, checkbox). Two-way binding
// via v-model, errors via `errors` array. For toggles, use DesignSwitch.

interface Option {
  value: string | number;
  label: string;
}

type InputType =
  | "text"
  | "email"
  | "password"
  | "number"
  | "url"
  | "tel"
  | "date"
  | "datetime-local"
  | "time"
  | "search"
  | "select"
  | "textarea"
  | "checkbox";

const props = withDefaults(
  defineProps<{
    type?: InputType;
    label?: string;
    errors?: string[];
    placeholder?: string;
    required?: boolean;
    disabled?: boolean;
    min?: number | string;
    max?: number | string;
    rows?: number;
    options?: Option[];
    prompt?: string;
    monospace?: boolean;
  }>(),
  {
    type: "text",
    errors: () => [],
    required: false,
    disabled: false,
    monospace: false,
  },
);

const model = defineModel<string | number | boolean | undefined>();

const fieldClass = (hasError: boolean) => [
  "w-full rounded-xl border bg-surface-default text-text-default text-body-2 px-3 py-2",
  props.monospace && "font-mono",
  hasError ? "border-semantic-error" : "border-border-1",
];
</script>

<template>
  <div class="mb-2">
    <!-- Checkbox: label sits inline with the input -->
    <template v-if="type === 'checkbox'">
      <label
        class="flex items-center gap-2 text-body-3 text-text-default cursor-pointer has-disabled:cursor-not-allowed has-disabled:opacity-50"
      >
        <input
          v-model="model as unknown as boolean"
          type="checkbox"
          :required="required"
          :disabled="disabled"
          class="size-4 rounded border border-border-1 accent-primary-contrast"
        />
        <span v-if="label">{{ label }}</span>
      </label>
    </template>

    <!-- All other types: label above field -->
    <label v-else class="block">
      <span v-if="label" class="block mb-1 text-title-3 text-text-default">
        {{ label }}
      </span>

      <select
        v-if="type === 'select'"
        v-model="model"
        :required="required"
        :disabled="disabled"
        :class="fieldClass(errors.length > 0)"
      >
        <option v-if="prompt" value="">{{ prompt }}</option>
        <option v-for="o in options" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>

      <textarea
        v-else-if="type === 'textarea'"
        v-model="model as unknown as string"
        :placeholder="placeholder"
        :required="required"
        :disabled="disabled"
        :rows="rows ?? 3"
        :class="fieldClass(errors.length > 0)"
      />

      <input
        v-else
        v-model="model"
        :type="type"
        :placeholder="placeholder"
        :required="required"
        :disabled="disabled"
        :min="min"
        :max="max"
        :class="fieldClass(errors.length > 0)"
      />
    </label>

    <p
      v-for="msg in errors"
      :key="msg"
      class="mt-1.5 flex items-center gap-2 text-body-3 text-semantic-error"
    >
      <Icon name="tabler:alert-circle" size="18" />
      {{ msg }}
    </p>
  </div>
</template>

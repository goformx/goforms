<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { type AlertVariants, alertVariants } from ".";
import { cn } from "@/lib/utils";

const props = defineProps<{
  variant?: AlertVariants["variant"];
  class?: HTMLAttributes["class"];
}>();
</script>

<template>
  <div
    role="alert"
    :class="cn(alertVariants({ variant }), props.class)"
  >
    <slot />
  </div>
</template>

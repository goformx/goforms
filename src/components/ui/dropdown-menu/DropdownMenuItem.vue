<script setup lang="ts">
import { computed, type HTMLAttributes } from "vue";
import { DropdownMenuItem, type DropdownMenuItemProps, useForwardProps } from "radix-vue";
import { cn } from "@/lib/utils";

const props = defineProps<DropdownMenuItemProps & { class?: HTMLAttributes["class"]; inset?: boolean }>();

const delegatedProps = computed(() => {
  const { class: _, inset: __, ...delegated } = props;
  return delegated;
});

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <DropdownMenuItem
    v-bind="forwardedProps"
    :class="
      cn(
        'relative flex cursor-default select-none items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&>svg]:size-4 [&>svg]:shrink-0',
        inset && 'pl-8',
        props.class,
      )
    "
  >
    <slot />
  </DropdownMenuItem>
</template>

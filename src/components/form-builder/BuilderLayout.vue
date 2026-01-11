<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { Logger } from "@/lib/core/logger";

interface Props {
  showFieldsPanel?: boolean;
  showSettingsPanel?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showFieldsPanel: true,
  showSettingsPanel: true,
});

// Panel collapse state (persisted in localStorage)
const isFieldsPanelCollapsed = ref(false);
const isSettingsPanelCollapsed = ref(false);

// Load panel state from localStorage
onMounted(() => {
  try {
    const savedFieldsState = localStorage.getItem("builder-fields-panel-collapsed");
    const savedSettingsState = localStorage.getItem("builder-settings-panel-collapsed");

    if (savedFieldsState !== null) {
      isFieldsPanelCollapsed.value = savedFieldsState === "true";
    }
    if (savedSettingsState !== null) {
      isSettingsPanelCollapsed.value = savedSettingsState === "true";
    }
  } catch (error) {
    Logger.error("Failed to load panel state:", error);
  }
});

/**
 * Toggle fields panel collapse state
 */
function toggleFieldsPanel() {
  isFieldsPanelCollapsed.value = !isFieldsPanelCollapsed.value;
  try {
    localStorage.setItem(
      "builder-fields-panel-collapsed",
      String(isFieldsPanelCollapsed.value)
    );
  } catch (error) {
    Logger.error("Failed to save fields panel state:", error);
  }
}

/**
 * Toggle settings panel collapse state
 */
function toggleSettingsPanel() {
  isSettingsPanelCollapsed.value = !isSettingsPanelCollapsed.value;
  try {
    localStorage.setItem(
      "builder-settings-panel-collapsed",
      String(isSettingsPanelCollapsed.value)
    );
  } catch (error) {
    Logger.error("Failed to save settings panel state:", error);
  }
}
</script>

<template>
  <div class="builder-layout flex flex-col h-full">
    <!-- Header Slot -->
    <div v-if="$slots['header']" class="builder-header border-b bg-background">
      <slot name="header" />
    </div>

    <!-- Three Panel Layout -->
    <div class="builder-content flex flex-1 overflow-hidden">
      <!-- Fields Panel (Left) -->
      <div
        v-if="props.showFieldsPanel"
        :class="[
          'builder-fields-panel border-r bg-background transition-all duration-200',
          isFieldsPanelCollapsed ? 'w-0' : 'w-64'
        ]"
      >
        <div v-if="!isFieldsPanelCollapsed" class="h-full flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b">
            <h3 class="font-semibold text-sm">Fields</h3>
            <Button
              variant="ghost"
              size="icon"
              class="h-6 w-6"
              @click="toggleFieldsPanel"
            >
              <ChevronLeft class="h-4 w-4" />
            </Button>
          </div>
          <div class="flex-1 overflow-hidden">
            <slot name="fields-panel" />
          </div>
        </div>
      </div>

      <!-- Collapsed Fields Panel Toggle -->
      <div
        v-if="props.showFieldsPanel && isFieldsPanelCollapsed"
        class="flex items-start pt-3 px-1 border-r"
      >
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6"
          @click="toggleFieldsPanel"
        >
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>

      <!-- Canvas (Center) -->
      <div class="builder-canvas flex-1 bg-muted/10 overflow-auto">
        <slot name="canvas" />
      </div>

      <!-- Collapsed Settings Panel Toggle -->
      <div
        v-if="props.showSettingsPanel && isSettingsPanelCollapsed"
        class="flex items-start pt-3 px-1 border-l"
      >
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6"
          @click="toggleSettingsPanel"
        >
          <ChevronLeft class="h-4 w-4" />
        </Button>
      </div>

      <!-- Settings Panel (Right) -->
      <div
        v-if="props.showSettingsPanel"
        :class="[
          'builder-settings-panel border-l bg-background transition-all duration-200',
          isSettingsPanelCollapsed ? 'w-0' : 'w-80'
        ]"
      >
        <div v-if="!isSettingsPanelCollapsed" class="h-full flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b">
            <h3 class="font-semibold text-sm">Settings</h3>
            <Button
              variant="ghost"
              size="icon"
              class="h-6 w-6"
              @click="toggleSettingsPanel"
            >
              <ChevronRight class="h-4 w-4" />
            </Button>
          </div>
          <div class="flex-1 overflow-hidden">
            <slot name="settings-panel" />
          </div>
        </div>
      </div>
    </div>

    <!-- Footer Slot -->
    <div v-if="$slots['footer']" class="builder-footer border-t bg-background">
      <slot name="footer" />
    </div>
  </div>
</template>

<style scoped>
.builder-layout {
  height: calc(100vh - 4rem); /* Adjust based on your header height */
}

/* Mobile responsive - stack panels vertically */
@media (max-width: 768px) {
  .builder-content {
    flex-direction: column;
  }

  .builder-fields-panel,
  .builder-settings-panel {
    width: 100% !important;
    height: auto;
    max-height: 40vh;
    border-right: none;
    border-left: none;
    border-bottom: 1px solid hsl(var(--border));
  }

  .builder-canvas {
    flex: 1;
  }
}
</style>

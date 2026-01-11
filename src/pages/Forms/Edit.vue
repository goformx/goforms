<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useForm, router, Link } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import BuilderLayout from "@/components/form-builder/BuilderLayout.vue";
import FieldSettingsPanel from "@/components/form-builder/FieldSettingsPanel.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useFormBuilder, type FormSchema } from "@/composables/useFormBuilder";
import { useKeyboardShortcuts, formatShortcut } from "@/composables/useKeyboardShortcuts";
import type { FormComponent } from "@/composables/useFormBuilderState";
import { Eye, ListChecks, Save, Code, Undo2, Redo2, Keyboard } from "lucide-vue-next";
import { toast } from "vue-sonner";

interface Form {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published" | "archived";
  corsOrigins: string[];
}

interface Props {
  form: Form;
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

// Form details form
const detailsForm = useForm({
  title: props.form.title,
  description: props.form.description,
  status: props.form.status,
  cors_origins: props.form.corsOrigins?.join(", ") ?? "",
});

const showSchemaModal = ref(false);
const showShortcutsModal = ref(false);
const isSavingAll = ref(false);

// Form builder with new features
const {
  isLoading: isBuilderLoading,
  error: builderError,
  isSaving,
  saveSchema,
  getSchema,
  selectedField,
  selectField,
  duplicateField,
  deleteField,
  undo,
  redo,
  canUndo,
  canRedo,
  exportSchema,
} = useFormBuilder({
  containerId: "form-schema-builder",
  formId: props.form.id,
  autoSave: false,
  onSchemaChange: (_schema: FormSchema) => {
    // Schema changed
  },
});

// Get selected field data
const selectedFieldData = computed<FormComponent | null>(() => {
  if (!selectedField.value) return null;

  const schema = getSchema();
  const findField = (components: unknown[]): FormComponent | null => {
    for (const comp of components) {
      const component = comp as FormComponent;
      if (component.key === selectedField.value) return component;
      if (component.components) {
        const found = findField(component.components as unknown[]);
        if (found) return found;
      }
    }
    return null;
  };

  return findField(schema.components);
});

// Keyboard shortcuts
const shortcuts = [
  {
    key: "s",
    meta: true,
    handler: () => void handleSave(),
    description: "Save form",
  },
  {
    key: "p",
    meta: true,
    handler: () => router.visit(`/forms/${props.form.id}/preview`),
    description: "Preview form",
  },
  {
    key: "z",
    meta: true,
    handler: () => undo(),
    description: "Undo",
  },
  {
    key: "z",
    meta: true,
    shift: true,
    handler: () => redo(),
    description: "Redo",
  },
  {
    key: "d",
    meta: true,
    handler: () => {
      if (selectedField.value) duplicateField(selectedField.value);
    },
    description: "Duplicate selected field",
  },
  {
    key: "Backspace",
    meta: true,
    handler: () => {
      if (selectedField.value) deleteField(selectedField.value);
    },
    description: "Delete selected field",
  },
  {
    key: "/",
    meta: true,
    handler: () => {
      showShortcutsModal.value = true;
    },
    description: "Show shortcuts",
  },
];

useKeyboardShortcuts(shortcuts);

async function handleSave() {
  if (isSavingAll.value || isSaving.value) return;

  // Validate CORS origins if publishing
  if (detailsForm.status === "published" && !detailsForm.cors_origins.trim()) {
    toast.error("CORS origins are required when publishing a form.");
    return;
  }

  isSavingAll.value = true;

  try {
    // Save schema first
    await saveSchema();

    // Then save details via Inertia
    await new Promise<void>((resolve, reject) => {
      detailsForm.put(`/forms/${props.form.id}`, {
        preserveScroll: true,
        onSuccess: () => resolve(),
        onError: () => reject(new Error("Failed to save form details")),
      });
    });

    toast.success("Form saved successfully");
  } catch (err) {
    const message = err instanceof Error ? err.message : "Failed to save form";
    toast.error(message);
  } finally {
    isSavingAll.value = false;
  }
}

function viewSchema() {
  showSchemaModal.value = true;
}

// Show flash messages reactively (when page re-renders with flash)
watch(
  () => props.flash,
  (flash) => {
    if (flash?.success) {
      toast.success(flash.success);
    }
    if (flash?.error) {
      toast.error(flash.error);
    }
  },
  { immediate: true }
);

// Show builder errors
watch(
  builderError,
  (error) => {
    if (error) {
      toast.error(error);
    }
  },
  { immediate: true }
);
</script>

<template>
  <DashboardLayout title="Edit Form" :subtitle="props.form.title">
    <template #actions>
      <div class="flex items-center gap-2">
        <!-- Status Badge -->
        <Badge :variant="props.form.status === 'published' ? 'default' : 'secondary'">
          {{ props.form.status }}
        </Badge>

        <Separator orientation="vertical" class="h-6" />

        <!-- Undo/Redo -->
        <Button variant="ghost" size="icon" :disabled="!canUndo" title="Undo (Cmd+Z)" @click="undo">
          <Undo2 class="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="icon" :disabled="!canRedo" title="Redo (Cmd+Shift+Z)" @click="redo">
          <Redo2 class="h-4 w-4" />
        </Button>

        <Separator orientation="vertical" class="h-6" />

        <!-- Actions -->
        <Button variant="outline" size="sm" @click="showShortcutsModal = true">
          <Keyboard class="mr-2 h-4 w-4" />
          Shortcuts
        </Button>
        <Button variant="outline" size="sm" @click="viewSchema">
          <Code class="mr-2 h-4 w-4" />
          Schema
        </Button>
        <Button variant="outline" size="sm" as-child>
          <Link :href="`/forms/${props.form.id}/preview`">
            <Eye class="mr-2 h-4 w-4" />
            Preview
          </Link>
        </Button>
        <Button variant="outline" size="sm" as-child>
          <Link :href="`/forms/${props.form.id}/submissions`">
            <ListChecks class="mr-2 h-4 w-4" />
            Submissions
          </Link>
        </Button>
        <Button size="sm" :disabled="isSavingAll || isBuilderLoading" @click="handleSave">
          <Save class="mr-2 h-4 w-4" />
          <span v-if="isSavingAll">Saving...</span>
          <span v-else>Save</span>
        </Button>
      </div>
    </template>

    <!-- Two-Panel Builder (Form.io provides its own sidebar) -->
    <BuilderLayout class="rounded-lg border bg-background shadow-sm" :show-fields-panel="false">
      <!-- Header Slot -->
      <template #header>
        <div class="px-6 py-4 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <!-- Form Title -->
            <div class="space-y-2">
              <Label for="title" class="text-xs">Form Title</Label>
              <Input id="title" v-model="detailsForm.title" type="text" placeholder="Enter form title" required />
            </div>

            <!-- Status -->
            <div class="space-y-2">
              <Label for="status" class="text-xs">Status</Label>
              <select id="status" v-model="detailsForm.status"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                <option value="draft">Draft</option>
                <option value="published">Published</option>
                <option value="archived">Archived</option>
              </select>
            </div>
          </div>

          <div class="space-y-2">
            <Label for="description" class="text-xs">Description</Label>
            <Input id="description" v-model="detailsForm.description" type="text"
              placeholder="Enter form description" />
          </div>

          <div class="space-y-2">
            <Label for="corsOrigins" class="text-xs">Allowed Origins (CORS)</Label>
            <Input id="corsOrigins" v-model="detailsForm.cors_origins" type="text"
              placeholder="e.g. *, https://example.com" />
            <p class="text-xs text-muted-foreground">
              Required when publishing. Use * to allow all origins.
            </p>
          </div>
        </div>
      </template>

      <!-- Canvas (Form.io renders its own sidebar here) -->
      <template #canvas>
        <div class="p-6">
          <div v-if="isBuilderLoading" class="flex items-center justify-center py-12">
            <div class="text-muted-foreground">Loading form builder...</div>
          </div>
          <div id="form-schema-builder" class="min-h-[500px]" :data-form-id="props.form.id" />
        </div>
      </template>

      <!-- Settings Panel -->
      <template #settings-panel>
        <FieldSettingsPanel :selected-field="selectedFieldData" @update:field="(field) => {
          // Update field in schema
          console.log('Update field:', field);
        }" @duplicate="(key) => duplicateField(key)" @delete="(key) => deleteField(key)"
          @close="() => selectField(null)" />
      </template>
    </BuilderLayout>

    <!-- Schema Modal -->
    <Dialog v-model:open="showSchemaModal">
      <DialogContent class="max-w-3xl max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>Form Schema (JSON)</DialogTitle>
        </DialogHeader>
        <div class="overflow-auto max-h-[60vh]">
          <pre class="text-xs bg-muted p-4 rounded-md overflow-auto">{{ exportSchema() }}</pre>
        </div>
        <div class="flex justify-end gap-2">
          <Button variant="outline" @click="showSchemaModal = false">Close</Button>
        </div>
      </DialogContent>
    </Dialog>

    <!-- Shortcuts Modal -->
    <Dialog v-model:open="showShortcutsModal">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>Keyboard Shortcuts</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <div v-for="shortcut in shortcuts" :key="shortcut.description" class="flex items-center justify-between py-2">
            <span class="text-sm">{{ shortcut.description }}</span>
            <kbd
              class="inline-flex items-center gap-1 rounded border border-border bg-muted px-2 py-1 text-xs font-mono">
              {{ formatShortcut(shortcut) }}
            </kbd>
          </div>
        </div>
        <div class="flex justify-end">
          <Button @click="showShortcutsModal = false">Close</Button>
        </div>
      </DialogContent>
    </Dialog>
  </DashboardLayout>
</template>

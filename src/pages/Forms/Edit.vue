<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useForm, router, Link } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useFormBuilder, type FormSchema } from "@/composables/useFormBuilder";
import { Eye, ListChecks, Save, Code, CheckCircle2, AlertCircle } from "lucide-vue-next";

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
  corsOrigins: props.form.corsOrigins?.join(", ") ?? "",
});

const successMessage = ref<string | null>(null);
const errorMessage = ref<string | null>(null);
const showSchemaModal = ref(false);

const isDetailsSubmitting = computed(() => detailsForm.processing);

// Form builder
const {
  isLoading: isBuilderLoading,
  error: builderError,
  isSaving,
  saveSchema,
  getSchema,
} = useFormBuilder({
  containerId: "form-schema-builder",
  formId: props.form.id,
  onSchemaChange: (_schema: FormSchema) => {
    // Optional: Auto-save or mark as dirty
  },
});

function handleDetailsSubmit() {
  successMessage.value = null;
  errorMessage.value = null;

  // Validate CORS origins if publishing
  if (detailsForm.status === "published" && !detailsForm.corsOrigins.trim()) {
    errorMessage.value = "CORS origins are required when publishing a form.";
    return;
  }

  detailsForm.put(`/forms/${props.form.id}`, {
    onSuccess: () => {
      successMessage.value = "Form details updated successfully";
    },
    onError: () => {
      errorMessage.value = "Failed to update form details";
    },
  });
}

async function handleSaveSchema() {
  successMessage.value = null;
  errorMessage.value = null;

  try {
    await saveSchema();
    successMessage.value = "Form schema saved successfully";
  } catch {
    errorMessage.value = "Failed to save form schema";
  }
}

function viewSchema() {
  showSchemaModal.value = true;
}

function closeSchemaModal() {
  showSchemaModal.value = false;
}

onMounted(() => {
  // Import Form.io styles
  const link = document.createElement("link");
  link.rel = "stylesheet";
  link.href = "/node_modules/@formio/js/dist/formio.full.min.css";
  document.head.appendChild(link);
});
</script>

<template>
  <DashboardLayout title="Edit Form" subtitle="Configure your form settings and fields">
    <template #actions>
      <Button variant="outline" as-child>
        <Link :href="`/forms/${props.form.id}/preview`">
          <Eye class="mr-2 h-4 w-4" />
          Preview
        </Link>
      </Button>
      <Button variant="outline" as-child>
        <Link :href="`/forms/${props.form.id}/submissions`">
          <ListChecks class="mr-2 h-4 w-4" />
          Submissions
        </Link>
      </Button>
    </template>

    <div class="space-y-6">
      <!-- Success/Error Messages -->
      <Alert v-if="successMessage || props.flash?.success" variant="success">
        <CheckCircle2 class="h-4 w-4" />
        <AlertDescription>{{ successMessage || props.flash?.success }}</AlertDescription>
      </Alert>

      <Alert v-if="errorMessage || builderError || props.flash?.error" variant="destructive">
        <AlertCircle class="h-4 w-4" />
        <AlertDescription>{{ errorMessage || builderError || props.flash?.error }}</AlertDescription>
      </Alert>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Form Details Sidebar -->
        <div class="lg:col-span-1">
          <Card>
            <CardHeader>
              <CardTitle>Form Details</CardTitle>
            </CardHeader>
            <form @submit.prevent="handleDetailsSubmit">
              <CardContent class="space-y-4">
                <div class="space-y-2">
                  <Label for="title">Form Title</Label>
                  <Input
                    id="title"
                    v-model="detailsForm.title"
                    type="text"
                    placeholder="Enter form title"
                    required
                  />
                </div>

                <div class="space-y-2">
                  <Label for="description">Description</Label>
                  <textarea
                    id="description"
                    v-model="detailsForm.description"
                    class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    placeholder="Enter form description"
                    rows="3"
                  />
                </div>

                <div class="space-y-2">
                  <Label for="status">Status</Label>
                  <select
                    id="status"
                    v-model="detailsForm.status"
                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    <option value="draft">Draft</option>
                    <option value="published">Published</option>
                    <option value="archived">Archived</option>
                  </select>
                </div>

                <div class="space-y-2">
                  <Label for="corsOrigins">Allowed Origins</Label>
                  <Input
                    id="corsOrigins"
                    v-model="detailsForm.corsOrigins"
                    type="text"
                    placeholder="e.g. *, https://example.com"
                  />
                  <p class="text-xs text-muted-foreground">
                    Required when publishing. Use * to allow all origins.
                  </p>
                </div>

                <div class="flex gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    @click="router.visit('/dashboard')"
                  >
                    Cancel
                  </Button>
                  <Button type="submit" :disabled="isDetailsSubmitting">
                    <span v-if="isDetailsSubmitting">Saving...</span>
                    <span v-else>Save Details</span>
                  </Button>
                </div>
              </CardContent>
            </form>
          </Card>
        </div>

        <!-- Form Builder -->
        <div class="lg:col-span-2">
          <Card>
            <CardHeader class="flex flex-row items-center justify-between">
              <CardTitle>Form Builder</CardTitle>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="viewSchema">
                  <Code class="mr-2 h-4 w-4" />
                  View Schema
                </Button>
                <Button size="sm" :disabled="isSaving" @click="handleSaveSchema">
                  <Save class="mr-2 h-4 w-4" />
                  <span v-if="isSaving">Saving...</span>
                  <span v-else>Save Fields</span>
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div v-if="isBuilderLoading" class="flex items-center justify-center py-12">
                <div class="text-muted-foreground">Loading form builder...</div>
              </div>
              <div
                id="form-schema-builder"
                class="min-h-[400px] border rounded-md"
                :data-form-id="props.form.id"
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>

    <!-- Schema Modal -->
    <div
      v-if="showSchemaModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="closeSchemaModal"
    >
      <div class="bg-background rounded-lg shadow-lg max-w-2xl w-full max-h-[80vh] overflow-hidden">
        <div class="flex items-center justify-between p-4 border-b">
          <h3 class="text-lg font-semibold">Form Schema (JSON)</h3>
          <Button variant="ghost" size="sm" @click="closeSchemaModal">
            &times;
          </Button>
        </div>
        <div class="p-4 overflow-auto max-h-[60vh]">
          <pre class="text-sm bg-muted p-4 rounded-md overflow-auto">{{ JSON.stringify(getSchema(), null, 2) }}</pre>
        </div>
        <div class="flex justify-end p-4 border-t">
          <Button @click="closeSchemaModal">Close</Button>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>

<style>
/* Form.io builder styles */
.formio-builder {
  background-color: var(--color-background);
}

.formio-component {
  margin-bottom: 1rem;
}
</style>

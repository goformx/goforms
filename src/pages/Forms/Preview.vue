<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Link } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { Pencil, CheckCircle2, AlertCircle } from "lucide-vue-next";

// Register GoFormX templates
Formio.use(goforms);

interface Form {
  id: string;
  title: string;
  description: string;
  status: string;
}

interface FormSchema {
  display?: string;
  components: unknown[];
}

interface Props {
  form: Form;
  schema?: FormSchema;
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const isLoading = ref(true);
const error = ref<string | null>(null);
const submissionSuccess = ref(false);

onMounted(async () => {
  const container = document.getElementById("form-preview");
  if (!container) {
    error.value = "Form preview container not found";
    isLoading.value = false;
    return;
  }

  try {
    let schema = props.schema;

    // Fetch schema if not provided
    if (!schema) {
      const response = await fetch(`/api/v1/forms/${props.form.id}/schema`);
      if (response.ok) {
        const data = await response.json();
        if (data.success && data.data) {
          schema = data.data;
        }
      }
    }

    if (!schema || !schema.components || schema.components.length === 0) {
      error.value = "This form has no fields yet. Add fields in the form builder.";
      isLoading.value = false;
      return;
    }

    // Create the form
    const form = await Formio.createForm(container, schema, {
      readOnly: false,
      noAlerts: true,
    });

    // Handle submission
    form.on("submit", async (submission: unknown) => {
      try {
        const response = await fetch(`/api/v1/public/forms/${props.form.id}/submit`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Requested-With": "XMLHttpRequest",
          },
          body: JSON.stringify(submission),
        });

        if (response.ok) {
          submissionSuccess.value = true;
        } else {
          error.value = "Failed to submit form";
        }
      } catch {
        error.value = "Failed to submit form";
      }
    });

    isLoading.value = false;
  } catch (err) {
    console.error("Failed to load form preview:", err);
    error.value = "Failed to load form preview";
    isLoading.value = false;
  }
});
</script>

<template>
  <DashboardLayout :title="`Preview: ${props.form.title}`" subtitle="Preview how your form will appear to users">
    <template #actions>
      <Button as-child>
        <Link :href="`/forms/${props.form.id}/edit`">
          <Pencil class="mr-2 h-4 w-4" />
          Edit Form
        </Link>
      </Button>
    </template>

    <div class="max-w-2xl mx-auto">
      <!-- Success Message -->
      <Alert v-if="submissionSuccess" variant="success" class="mb-6">
        <CheckCircle2 class="h-4 w-4" />
        <AlertDescription>
          Form submitted successfully! This is a preview - no data was actually saved.
        </AlertDescription>
      </Alert>

      <!-- Error Message -->
      <Alert v-if="error" variant="destructive" class="mb-6">
        <AlertCircle class="h-4 w-4" />
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <Card class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle>{{ props.form.title }}</CardTitle>
          <p v-if="props.form.description" class="text-muted-foreground">
            {{ props.form.description }}
          </p>
        </CardHeader>
        <CardContent>
          <div v-if="isLoading" class="flex items-center justify-center py-12">
            <div class="text-muted-foreground">Loading form...</div>
          </div>
          <div id="form-preview" class="min-h-[200px]" />
        </CardContent>
      </Card>
    </div>
  </DashboardLayout>
</template>

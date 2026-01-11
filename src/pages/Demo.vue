<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Link } from "@inertiajs/vue3";
import GuestLayout from "@/components/layout/GuestLayout.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { CheckCircle2 } from "lucide-vue-next";

// Register GoFormX templates
Formio.use(goforms);

const isLoading = ref(true);
const error = ref<string | null>(null);
const submissionSuccess = ref(false);

// Demo form schema
const demoSchema = {
  display: "form",
  components: [
    {
      type: "textfield",
      key: "name",
      label: "Full Name",
      placeholder: "Enter your full name",
      validate: {
        required: true,
      },
    },
    {
      type: "email",
      key: "email",
      label: "Email Address",
      placeholder: "Enter your email",
      validate: {
        required: true,
      },
    },
    {
      type: "textarea",
      key: "message",
      label: "Message",
      placeholder: "Enter your message",
      rows: 4,
      validate: {
        required: true,
      },
    },
    {
      type: "select",
      key: "interest",
      label: "What are you interested in?",
      data: {
        values: [
          { label: "Self-hosting", value: "selfhosting" },
          { label: "API Integration", value: "api" },
          { label: "Form Builder", value: "builder" },
          { label: "Other", value: "other" },
        ],
      },
      validate: {
        required: true,
      },
    },
    {
      type: "checkbox",
      key: "subscribe",
      label: "Subscribe to updates",
      defaultValue: true,
    },
    {
      type: "button",
      action: "submit",
      label: "Submit",
      theme: "primary",
      block: true,
    },
  ],
};

onMounted(async () => {
  const container = document.getElementById("demo-form");
  if (!container) {
    error.value = "Demo form container not found";
    isLoading.value = false;
    return;
  }

  try {
    const form = await Formio.createForm(container, demoSchema, {
      noAlerts: true,
    });

    form.on("submit", () => {
      submissionSuccess.value = true;
      // Reset after showing success
      setTimeout(() => {
        submissionSuccess.value = false;
        form.resetValue();
      }, 3000);
    });

    isLoading.value = false;
  } catch (err) {
    console.error("Failed to load demo form:", err);
    error.value = "Failed to load demo form";
    isLoading.value = false;
  }
});
</script>

<template>
  <GuestLayout title="Demo">
    <div class="container py-12">
      <div class="max-w-2xl mx-auto">
        <!-- Header -->
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold tracking-tight sm:text-4xl">
            Try the Form Builder
          </h1>
          <p class="mt-4 text-lg text-muted-foreground">
            Experience the power of GoFormX with this interactive demo. Fill out the form below to see it in action.
          </p>
        </div>

        <!-- Demo Form Card -->
        <Card>
          <CardHeader>
            <CardTitle>Contact Form Demo</CardTitle>
            <CardDescription>
              This is a sample contact form built with GoFormX.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <!-- Success Message -->
            <Alert v-if="submissionSuccess" variant="success" class="mb-6">
              <CheckCircle2 class="h-4 w-4" />
              <AlertDescription>
                Form submitted successfully! This is a demo - no data was saved.
              </AlertDescription>
            </Alert>

            <!-- Loading State -->
            <div v-if="isLoading" class="flex items-center justify-center py-12">
              <div class="text-muted-foreground">Loading demo form...</div>
            </div>

            <!-- Error State -->
            <div v-else-if="error" class="text-center py-12">
              <p class="text-destructive">{{ error }}</p>
            </div>

            <!-- Form Container -->
            <div id="demo-form" class="min-h-[200px]" />
          </CardContent>
        </Card>

        <!-- CTA -->
        <div class="mt-8 text-center">
          <p class="text-muted-foreground mb-4">
            Ready to build your own forms?
          </p>
          <div class="flex justify-center gap-4">
            <Button as-child>
              <Link href="/signup">Get Started Free</Link>
            </Button>
            <Button variant="outline" as-child>
              <a
                href="https://github.com/goformx/goforms"
                target="_blank"
                rel="noopener noreferrer"
              >
                View on GitHub
              </a>
            </Button>
          </div>
        </div>
      </div>
    </div>
  </GuestLayout>
</template>

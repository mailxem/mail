"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { Form } from "@/app/types";
import { useTeam } from "./team-provider";
import { useApi } from "@/hooks/use-api";

type FormsContextType = {
  forms: Form[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
};

const FormsContext = createContext<FormsContextType>({
  forms: [],
  isLoading: true,
  error: null,
  refetch: async () => {},
});

export function useForms() {
  return useContext(FormsContext);
}

export function FormsProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [forms, setForms] = useState<Form[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const { team } = useTeam();
  const { apiFetch } = useApi();

  const fetchForms = async () => {
    try {
      setIsLoading(true);
      const response = await apiFetch("forms?teamId=" + team?.id);
      if (!response.ok) {
        throw new Error("Failed to fetch forms");
      }
      const { data } = await response.json();
      setForms(data);
      setError(null);
    } catch (err) {
      setError(
        err instanceof Error ? err : new Error("Unknown error occurred")
      );
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (team?.id) {
      fetchForms();
    }
  }, [team?.id]);

  return (
    <FormsContext.Provider
      value={{
        forms,
        isLoading,
        error,
        refetch: fetchForms,
      }}
    >
      {children}
    </FormsContext.Provider>
  );
} 
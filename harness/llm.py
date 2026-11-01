"""
Capa de Abstracción de Modelos (LLM Gateway).
Permite conectar el arnés a cualquier proveedor de IA de forma totalmente agnóstica.
"""

import os
from typing import Any
from dotenv import load_dotenv

load_dotenv()


def get_configured_llm(provider: str = None, model: str = None) -> Any:
    """
    Instancia el modelo de chat correspondiente según la configuración.
    Soporta: gemini, openai, anthropic, deepseek, ollama o modo simulado.
    """
    provider = provider or os.getenv("AI_PROVIDER", "gemini").lower()
    model = model or os.getenv("AI_MODEL", "")

    if provider == "gemini":
        from langchain_google_genai import ChatGoogleGenerativeAI
        api_key = os.getenv("GEMINI_API_KEY")
        model_name = model or "gemini-2.5-flash"
        return ChatGoogleGenerativeAI(
            model=model_name,
            google_api_key=api_key,
            temperature=0.1
        )

    elif provider in ("openai", "deepseek", "openrouter"):
        from langchain_openai import ChatOpenAI
        if provider == "deepseek":
            api_key = os.getenv("DEEPSEEK_API_KEY") or os.getenv("OPENROUTER_API_KEY")
            base_url = "https://api.deepseek.com" if os.getenv("DEEPSEEK_API_KEY") else "https://openrouter.ai/api/v1"
            model_name = model or "deepseek-chat"
            return ChatOpenAI(
                model=model_name,
                openai_api_key=api_key,
                openai_api_base=base_url,
                temperature=0.1
            )
        elif provider == "openrouter":
            api_key = os.getenv("OPENROUTER_API_KEY")
            model_name = model or "deepseek/deepseek-chat"
            return ChatOpenAI(
                model=model_name,
                openai_api_key=api_key,
                openai_api_base="https://openrouter.ai/api/v1",
                temperature=0.1
            )
        else:
            api_key = os.getenv("OPENAI_API_KEY")
            model_name = model or "gpt-4o"
            return ChatOpenAI(
                model=model_name,
                openai_api_key=api_key,
                temperature=0.1
            )

    elif provider == "anthropic":
        from langchain_anthropic import ChatAnthropic
        api_key = os.getenv("ANTHROPIC_API_KEY")
        model_name = model or "claude-3-7-sonnet-20250219"
        return ChatAnthropic(
            model=model_name,
            anthropic_api_key=api_key,
            temperature=0.1
        )

    elif provider == "ollama":
        from langchain_openai import ChatOpenAI
        base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434/v1")
        model_name = model or "qwen2.5-coder:14b"
        return ChatOpenAI(
            model=model_name,
            openai_api_key="ollama",
            openai_api_base=base_url,
            temperature=0.1
        )

    else:
        raise ValueError(f"Proveedor de IA no reconocido: '{provider}'. Opciones válidas: gemini, openai, anthropic, deepseek, openrouter, ollama.")

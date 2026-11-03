import { AxiosError, AxiosHeaders, type AxiosResponse } from "axios";
import { beforeEach, describe, expect, it, vi } from "vitest";

const put = vi.fn();
vi.mock("@/services/api/client", () => ({
  api: { put: (...args: unknown[]) => put(...args) },
}));

const { httpGateway } = await import("./httpGateway");

const postulacion = { id: "p1", estado: "NO_ELEGIBLE", version: 4 };
const elegibilidad = {
  elegible: false,
  cargoId: 2,
  mensaje: "Mensaje institucional",
  fecha: "2026-11-03T15:04:05Z",
};

function errorHttp(status: number, data: unknown): AxiosError {
  const config = { headers: new AxiosHeaders() };
  const response = {
    status,
    data,
    statusText: "",
    headers: {},
    config,
  } as AxiosResponse;
  return new AxiosError("fallo", String(status), config, undefined, response);
}

describe("httpGateway.confirmarCargo (SRP-017a, SDD 7.5.4)", () => {
  beforeEach(() => {
    put.mockClear();
  });

  it("returns the eligibility of a 200 response", async () => {
    put.mockResolvedValue({
      data: {
        success: true,
        data: {
          postulacion: { ...postulacion, estado: "BORRADOR" },
          elegibilidad: { ...elegibilidad, elegible: true },
        },
      },
    });
    const r = await httpGateway.confirmarCargo("p1", 1, 3);
    expect(put).toHaveBeenCalledWith("/postulaciones/p1/cargo", {
      cargoId: 1,
      confirmacion: true,
      version: 3,
    });
    expect(r.elegibilidad.elegible).toBe(true);
  });

  it("turns 422 POS_NO_ELEGIBLE_* into a not eligible result", async () => {
    put.mockImplementation(() =>
      Promise.reject(
        errorHttp(422, {
          success: false,
          code: "POS_NO_ELEGIBLE_JUB",
          data: { postulacion, elegibilidad },
        }),
      ),
    );
    const r = await httpGateway.confirmarCargo("p1", 2, 3);
    expect(r.postulacion.estado).toBe("NO_ELEGIBLE");
    expect(r.elegibilidad).toEqual(elegibilidad);
  });

  it("propagates 503 so the section offers a retry", async () => {
    put.mockImplementation(() =>
      Promise.reject(
        errorHttp(503, { success: false, code: "INTEGRACION_NO_DISPONIBLE" }),
      ),
    );
    await expect(httpGateway.confirmarCargo("p1", 1, 3)).rejects.toBeInstanceOf(
      AxiosError,
    );
  });

  it("propagates a 422 that is not an eligibility rejection", async () => {
    put.mockImplementation(() =>
      Promise.reject(
        errorHttp(422, { success: false, code: "VALIDATION_ERROR" }),
      ),
    );
    await expect(httpGateway.confirmarCargo("p1", 1, 3)).rejects.toBeInstanceOf(
      AxiosError,
    );
  });
});

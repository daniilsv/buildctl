import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { setToken } from "../utils/auth";

export default function Callback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const code = searchParams.get("code");
    const state = searchParams.get("state");

    if (!code || !state) {
      navigate("/auth/login");
      return;
    }

    // Exchange code for token via backend callback
    // Use fetch directly to avoid interceptor issues
    fetch(
      `/auth/callback?code=${encodeURIComponent(
        code
      )}&state=${encodeURIComponent(state)}`,
      {
        credentials: "include",
      }
    )
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
      })
      .then((data) => {
        console.log("Callback response:", data);
        const { access_token } = data;
        if (access_token) {
          setToken(access_token);
          console.log("Token saved, redirecting to home");
          navigate("/");
        } else {
          console.error("No access_token in response");
          navigate("/auth/login");
        }
      })
      .catch((error) => {
        console.error("Callback error:", error);
        navigate("/auth/login");
      });
  }, [searchParams, navigate]);

  return (
    <div
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "100vh",
      }}
    >
      <div>Authenticating...</div>
    </div>
  );
}

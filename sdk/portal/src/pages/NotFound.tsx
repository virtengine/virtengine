import { Link } from "react-router-dom";

import { Button } from "../components/ui/button";

export const NotFound = () => {
  return (
    <div className="flex min-h-[60vh] flex-col items-start justify-center gap-4">
      <h2 className="text-3xl font-semibold text-slate-100">Page not found</h2>
      <p className="text-sm text-slate-400">
        The page you requested doesn't exist.
      </p>
      <Button asChild>
        <Link to="/">Back to overview</Link>
      </Button>
    </div>
  );
};

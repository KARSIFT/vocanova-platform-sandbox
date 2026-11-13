import type { ReactNode } from "react";

/** A shared reading order for the existing learning activities. */
export function LearningIntro({
  title,
  eyebrow,
  description,
  headingLevel = 2,
  id,
  className,
}: {
  title: ReactNode;
  eyebrow?: string;
  description?: ReactNode;
  headingLevel?: 1 | 2 | 3;
  id?: string;
  className?: string;
}): ReactNode {
  const Heading = `h${headingLevel}` as "h1" | "h2" | "h3";
  return (
    <div className={["learning-intro", className].filter(Boolean).join(" ")}>
      {eyebrow ? <p className="learning-eyebrow">{eyebrow}</p> : null}
      <Heading id={id} className="learning-title">
        {title}
      </Heading>
      {description ? (
        <div className="learning-description">{description}</div>
      ) : null}
    </div>
  );
}

/** Guidance only: confirmation and live announcements stay with their owners. */
export function LearningNote({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}): ReactNode {
  return (
    <div className={["learning-note", className].filter(Boolean).join(" ")}>
      {children}
    </div>
  );
}

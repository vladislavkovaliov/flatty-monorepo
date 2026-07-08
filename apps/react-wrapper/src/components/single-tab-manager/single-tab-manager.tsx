import { useEffect, useRef, useState, type PropsWithChildren } from "react";
import {
  SingleTabManager,
  type SingleTabManagerOptions,
} from "single-active-browser-tab";

export const SingleTabManagerWrapper = ({ children }: PropsWithChildren) => {
  const [isActive, setIsActive] = useState<boolean | null>(null);

  const managerRef = useRef<SingleTabManager | null>(null);

  useEffect(() => {
    const singleTabManager = new SingleTabManager("broadcast", {
      onActive: () => {
        setIsActive(true);
      },
      onBlocked: () => {
        setIsActive(false);
      },
      logLevel: "log",
    } satisfies SingleTabManagerOptions);

    managerRef.current = singleTabManager;

    singleTabManager.start();

    return () => {
      singleTabManager.stop();
      managerRef.current = null;
    };
  }, []);

  const handleReloadCallback = () => {
    managerRef.current?.takeover();
  };

  return (
    <>
      {isActive === null && null}
      {isActive === true && children}
      {isActive === false && (
        <div>
          <p>Application is already opened in other tabs.</p>
          <button onClick={handleReloadCallback}>Reload</button>
        </div>
      )}
    </>
  );
};

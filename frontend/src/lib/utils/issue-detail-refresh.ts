import { browser } from "$app/environment";

const ISSUE_DETAIL_REFRESH_EVENT = "jedug:issue-detail-refresh";

export type IssueDetailRefreshDetail = {
  issueID: string;
  source?: "notification";
};

export function requestIssueDetailRefresh(detail: IssueDetailRefreshDetail) {
  if (!browser) return;

  window.dispatchEvent(
    new CustomEvent<IssueDetailRefreshDetail>(ISSUE_DETAIL_REFRESH_EVENT, {
      detail,
    }),
  );
}

export function onIssueDetailRefresh(
  listener: (detail: IssueDetailRefreshDetail) => void,
) {
  if (!browser) {
    return () => {};
  }

  const handler = (event: Event) => {
    const customEvent = event as CustomEvent<IssueDetailRefreshDetail>;
    if (!customEvent.detail?.issueID) return;
    listener(customEvent.detail);
  };

  window.addEventListener(ISSUE_DETAIL_REFRESH_EVENT, handler);

  return () => {
    window.removeEventListener(ISSUE_DETAIL_REFRESH_EVENT, handler);
  };
}

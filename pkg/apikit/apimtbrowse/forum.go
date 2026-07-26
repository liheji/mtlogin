package apimtbrowse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"mtlogin/pkg/log"
)

func (s *Session) BrowseForum(ctx *log.Context) (bool, error) {
	ctx.Info("forum: starting forum browsing")

	fids, authorIDs, err := s.PhaseVisitForums(ctx)
	if err != nil {
		if shouldPropagateForumError(err) {
			return false, err
		}
		ctx.Warnf("forum: phase 1 (forums) failed: %v", err)
		return false, nil
	}
	if len(fids) == 0 {
		ctx.Warn("forum: no forums found, skipping")
		return false, nil
	}

	if len(authorIDs) > 0 {
		if err := s.PhaseResolveMembers(ctx, authorIDs); err != nil {
			if shouldPropagateForumError(err) {
				return false, err
			}
			ctx.Warnf("forum: phase 1.5 (resolve forum authors) failed: %v", err)
		}
	}

	lastBrowseUpdated := false
	fid := fids[rand.Intn(len(fids))]
	ctx.Infof("forum: phase 2 - visiting random forum fid=%d", fid)

	if err := s.Warmup(ctx); err != nil {
		if shouldPropagateForumError(err) {
			return lastBrowseUpdated, err
		}
		ctx.Warnf("forum: phase 2 warmup failed: %v", err)
	}
	if err := s.UpdateLastBrowse(ctx); err != nil {
		if shouldPropagateForumError(err) {
			return lastBrowseUpdated, err
		}
		ctx.Warnf("forum: phase 2 updateLastBrowse failed: %v", err)
	} else {
		lastBrowseUpdated = true
	}

	tids, topicAuthorIDs, lastPostAuthorIDs, err := s.PhaseSearchTopics(ctx, fid)
	if err != nil {
		if shouldPropagateForumError(err) {
			return lastBrowseUpdated, err
		}
		ctx.Warnf("forum: phase 2 (topic search) failed: %v", err)
		return lastBrowseUpdated, nil
	}
	if len(tids) == 0 {
		ctx.Info("forum: no topics found, skipping phase 3")
		return lastBrowseUpdated, nil
	}

	if len(topicAuthorIDs) > 0 {
		if err := s.PhaseResolveMembers(ctx, topicAuthorIDs); err != nil {
			if shouldPropagateForumError(err) {
				return lastBrowseUpdated, err
			}
			ctx.Warnf("forum: phase 2.5 (resolve topic authors) failed: %v", err)
		}
	}
	if len(lastPostAuthorIDs) > 0 {
		if err := s.PhaseResolveMembers(ctx, lastPostAuthorIDs); err != nil {
			if shouldPropagateForumError(err) {
				return lastBrowseUpdated, err
			}
			ctx.Warnf("forum: phase 2.5 (resolve last posters) failed: %v", err)
		}
	}

	tid := tids[rand.Intn(len(tids))]
	ctx.Infof("forum: phase 3 - visiting topic tid=%d", tid)

	if err := s.Warmup(ctx); err != nil {
		if shouldPropagateForumError(err) {
			return lastBrowseUpdated, err
		}
		ctx.Warnf("forum: phase 3 warmup failed: %v", err)
	}
	if err := s.UpdateLastBrowse(ctx); err != nil {
		if shouldPropagateForumError(err) {
			return lastBrowseUpdated, err
		}
		ctx.Warnf("forum: phase 3 updateLastBrowse failed: %v", err)
	} else {
		lastBrowseUpdated = true
	}

	if err := s.PhaseViewTopic(ctx, tid); err != nil {
		if shouldPropagateForumError(err) {
			return lastBrowseUpdated, err
		}
		ctx.Warnf("forum: phase 3 (topic view) failed: %v", err)
		return lastBrowseUpdated, nil
	}

	ctx.Info("forum: browsing completed successfully")
	return lastBrowseUpdated, nil
}

func shouldPropagateForumError(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, ErrAuthFailed)
}

func (s *Session) PhaseVisitForums(ctx *log.Context) (fids, authorIDs []int64, err error) {
	resp, err := s.do(ctx, http.MethodPost, "/api/forum/forums", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("forums request: %w", err)
	}
	var data ForumsData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, nil, fmt.Errorf("%w: forums data decode: %v", ErrBusiness, err)
	}

	for _, f := range data.ForumsList {
		if f.ID > 0 {
			fids = append(fids, f.ID)
		}
	}

	seen := make(map[int64]bool)
	for _, lp := range data.LastPost {
		if aid := lp.Post.AuthorID; aid > 0 && !seen[aid] {
			seen[aid] = true
			authorIDs = append(authorIDs, aid)
		}
	}

	ctx.Infof("forum: phase 1 done forums=%d authors=%d", len(fids), len(authorIDs))
	return fids, authorIDs, nil
}

func (s *Session) PhaseResolveMembers(ctx *log.Context, ids []int64) error {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.FormatInt(id, 10)
	}

	resp, err := s.doJSON(ctx, http.MethodPost, "/api/member/bases", map[string]any{
		"ids": idStrs,
	})
	if err != nil {
		return fmt.Errorf("member bases request: %w", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("%w: member bases data decode: %v", ErrBusiness, err)
	}
	ctx.Infof("forum: resolved members count=%d", len(data))
	return nil
}

func (s *Session) PhaseSearchTopics(ctx *log.Context, fid int64) (tids, authorIDs, lastPostAuthorIDs []int64, err error) {
	resp, err := s.doJSON(ctx, http.MethodPost, "/api/forum/topic/search", map[string]any{
		"fid":        strconv.FormatInt(fid, 10),
		"pageSize":   100,
		"pageNumber": 1,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("topic search request: %w", err)
	}
	var data TopicSearchData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, nil, nil, fmt.Errorf("%w: topic search data decode: %v", ErrBusiness, err)
	}

	seenAuthors := make(map[int64]bool)
	seenPosters := make(map[int64]bool)
	for _, item := range data.Data {
		if item.ID > 0 {
			tids = append(tids, item.ID)
		}
		if item.Author > 0 && !seenAuthors[item.Author] {
			seenAuthors[item.Author] = true
			authorIDs = append(authorIDs, item.Author)
		}
		if lpa := item.LastPost.AuthorID; lpa > 0 && !seenPosters[lpa] {
			seenPosters[lpa] = true
			lastPostAuthorIDs = append(lastPostAuthorIDs, lpa)
		}
	}

	ctx.Infof("forum: phase 2 done topics=%d authors=%d posters=%d", len(tids), len(authorIDs), len(lastPostAuthorIDs))
	return tids, authorIDs, lastPostAuthorIDs, nil
}

func (s *Session) PhaseViewTopic(ctx *log.Context, tid int64) error {
	tidStr := strconv.FormatInt(tid, 10)

	if _, err := s.do(ctx, http.MethodPost, "/api/forum/topic/viewHits", map[string]string{
		"tid": tidStr,
	}); err != nil {
		return fmt.Errorf("viewHits request: %w", err)
	}
	if _, err := s.do(ctx, http.MethodPost, "/api/forum/forums", nil); err != nil {
		return fmt.Errorf("forums request: %w", err)
	}
	if _, err := s.doJSON(ctx, http.MethodPost, "/api/forum/topic/detail", map[string]any{
		"tid":        tidStr,
		"pageSize":   20,
		"pageNumber": 1,
	}); err != nil {
		return fmt.Errorf("topic detail request: %w", err)
	}

	ctx.Infof("forum: phase 3 done tid=%d", tid)
	return nil
}

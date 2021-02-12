"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = require("@aws-cdk/assert");
const cdk = require("@aws-cdk/core");
//import * as ImageRecog from '../lib/image-recog-stack';
const ImageRecog = require("../setup");
test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new ImageRecog.ImageRecogStack(app, 'MyTestStack');
    // THEN
    assert_1.expect(stack).to(assert_1.matchTemplate({
        "Resources": {}
    }, assert_1.MatchStyle.EXACT));
});
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW1hZ2UtcmVjb2cudGVzdC5qcyIsInNvdXJjZVJvb3QiOiIiLCJzb3VyY2VzIjpbIi4uLy4uL3Rlc3QvaW1hZ2UtcmVjb2cudGVzdC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOztBQUFBLDRDQUFpRjtBQUNqRixxQ0FBcUM7QUFDckMseURBQXlEO0FBQ3pELHVDQUF1QztBQUV2QyxJQUFJLENBQUMsYUFBYSxFQUFFLEdBQUcsRUFBRTtJQUN2QixNQUFNLEdBQUcsR0FBRyxJQUFJLEdBQUcsQ0FBQyxHQUFHLEVBQUUsQ0FBQztJQUMxQixPQUFPO0lBQ1AsTUFBTSxLQUFLLEdBQUcsSUFBSSxVQUFVLENBQUMsZUFBZSxDQUFDLEdBQUcsRUFBRSxhQUFhLENBQUMsQ0FBQztJQUNqRSxPQUFPO0lBQ1AsZUFBUyxDQUFDLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxzQkFBYSxDQUFDO1FBQ2hDLFdBQVcsRUFBRSxFQUFFO0tBQ2hCLEVBQUUsbUJBQVUsQ0FBQyxLQUFLLENBQUMsQ0FBQyxDQUFBO0FBQ3ZCLENBQUMsQ0FBQyxDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgZXhwZWN0IGFzIGV4cGVjdENESywgbWF0Y2hUZW1wbGF0ZSwgTWF0Y2hTdHlsZSB9IGZyb20gJ0Bhd3MtY2RrL2Fzc2VydCc7XHJcbmltcG9ydCAqIGFzIGNkayBmcm9tICdAYXdzLWNkay9jb3JlJztcclxuLy9pbXBvcnQgKiBhcyBJbWFnZVJlY29nIGZyb20gJy4uL2xpYi9pbWFnZS1yZWNvZy1zdGFjayc7XHJcbmltcG9ydCAqIGFzIEltYWdlUmVjb2cgZnJvbSAnLi4vc2V0dXAnO1xyXG5cclxudGVzdCgnRW1wdHkgU3RhY2snLCAoKSA9PiB7XHJcbiAgY29uc3QgYXBwID0gbmV3IGNkay5BcHAoKTtcclxuICAvLyBXSEVOXHJcbiAgY29uc3Qgc3RhY2sgPSBuZXcgSW1hZ2VSZWNvZy5JbWFnZVJlY29nU3RhY2soYXBwLCAnTXlUZXN0U3RhY2snKTtcclxuICAvLyBUSEVOXHJcbiAgZXhwZWN0Q0RLKHN0YWNrKS50byhtYXRjaFRlbXBsYXRlKHtcclxuICAgIFwiUmVzb3VyY2VzXCI6IHt9XHJcbiAgfSwgTWF0Y2hTdHlsZS5FWEFDVCkpXHJcbn0pO1xyXG4iXX0=